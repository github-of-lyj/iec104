package iec104

import (
	"encoding/binary"
	"fmt"
)

/*
ASDU (Application Service Data Unit).

The ASDU contains two main sections:
- the data unit identifier (with the fixed length of six bytes):
  - defining the specific type of data;
  - providing addressing to identify the specific data;
  - including information as cause of transmission.

- the data itself, made up of one or more information objects:
  - each ASDU can transmit maximum 127 objects;
  - the type identification is applied to the entire ASDU, so the information objects contained in the ASDU
    are of the same type.

The format of ASDU:

	| <-              8 bits              -> |
	| Type Identification                    |  --------------------
	| SQ | Number of objects                 |           |
	| T  | P/N | Cause of transmission (COT) |           |
	| Original address (ORG)                 |  Data Uint Identifier
	| ASDU address fields                    |           |
	| ASDU address fields                    |  --------------------
	| Information object address (IOA)       |  --------------------
	| Information object address (IOA)       |           |
	| Information object address (IOA)       |  Information Object 1
	| Information Elements                   |           |
	| Time Tag                               |  --------------------
	| Information Object 2                   |
	| Information Object N                   |
*/
type ASDU struct {
	// Data Uint Identifier(with the fixed length of 6 bytes)
	typeID TypeID // 8  bits
	sq     SQ     // 1  bit
	nObjs  NOO    // 7  bits
	t      T      // 1  bit
	pn     PN     // 1  bit
	cot    COT    // 6  bits
	org    ORG    // 8  bits
	coa    COA    // 16 bits

	toBeHandled bool
	sendSFrame  bool
	cmdRsp      *cmdRsp

	ios     []*InformationObject
	Signals []*InformationElement
}

func (asdu *ASDU) Parse(data []byte) error {
	// I-format frame have ASDU.
	if len(data) < AsduHeaderLen {
		return fmt.Errorf("invalid asdu header: % X", data)
	}

	// the 1st byte
	asdu.parseTypeID(data[0])
	// the 2nd byte
	asdu.parseSQ(data[1])
	asdu.parseNOO(data[1])
	// the 3rd byte
	asdu.parseT(data[2])
	asdu.parsePN(data[2])
	asdu.parseCOT(data[2])
	// the 4th byte
	asdu.parseORG(data[3])
	// the 5th and 6th bytes
	asdu.parseCOA(data[4:AsduHeaderLen])

	asdu.parseInformationObjects(data[AsduHeaderLen:])
	return nil
}

func (asdu *ASDU) Data() []byte {
	data := make([]byte, 0)
	// the 1st byte
	data = append(data, byte(asdu.typeID))
	// the 2nd byte
	data = append(data, func() byte {
		if asdu.sq {
			return (0b1 << 7) & asdu.nObjs
		} else {
			return asdu.nObjs
		}
	}())
	// the 3rd byte
	data = append(data, func() byte {
		if bool(asdu.t) && bool(asdu.pn) {
			return (0b11 << 6) & byte(asdu.cot)
		} else if asdu.t {
			return (0b1 << 7) & byte(asdu.cot)
		} else if asdu.pn {
			return (0b1 << 6) & byte(asdu.cot)
		} else {
			return byte(asdu.cot)
		}
	}())
	// the 4th byte
	data = append(data, byte(asdu.org))
	// the 5th and 6th bytes
	data = append(data, func() []byte {
		x := make([]byte, 2, 2)
		binary.LittleEndian.PutUint16(x, asdu.coa)
		return x
	}()...)

	// the remaining bytes (some information objects)
	data = append(data, func() []byte {
		x := make([]byte, 0)
		for _, signal := range asdu.ios {
			x = append(x, signal.Data()...)
		}
		return x
	}()...)
	return data
}

/*
TypeID (Type Identification, 1 byte):
- value range:
  - 0 is not used;
  - 1-127 is used for standard IEC 101 definitions, there are presently 58 specific types defined:
    | Type ID | Group                                    |
    | 1-40    | Process information in monitor direction |
    | 45-51   | Process information in control direction |
    | 70      | System information in monitor direction  |
    | 100-106 | System information in control direction  |
    | 110-113 | Parameter in control direction           |
    | 120-126 | File transfer                            |
  - 128-135 is reserved for message routing;
  - 136-255 for special use.
*/
type TypeID uint8

const (
	// Process information in monitor direction

	// MSpNa1 indicates single point information.
	// InformationElementType: SIQ
	// COT: CotPerCyc, CotSpont, CotInrogen
	// [遥信 - 单点 - 不带时标]
	MSpNa1 TypeID = 0x1 // 1
	// MSpTa1 indicates single point information with time tag CP24Time2a.
	// InformationElementType: SIQ + CP24Time2a
	// COT: CotSpont
	// [遥信 - 单点 - 三字节时标]
	MSpTa1 TypeID = 0x2 // 2
	// MDpNa1 indicates double point information.
	// InformationElementType: DIQ
	// COT: CotSpont
	// [遥信 - 双点 - 不带时标]
	MDpNa1 TypeID = 0x3 // 3
	// MDpTa1 indicates double point information with time tag CP24Time2a.
	// InformationElementType: DIQ + CP24Time2a
	// COT: CotSpont
	// [遥信 - 双点 - 三字节时标]
	MDpTa1 TypeID = 0x4 // 4
	// MMeNa1 indicates measured value, normalized value.
	// InformationElementType: NVA + QDS
	// COT: 2, 3, 5, 11, 12, 20, 20+G
	MMeNa1 TypeID = 0x9 // 9
	// MMeTa1 indicates measured value, normalized value with time tag CP24Time2a.
	// InformationElementType: NVA + QDS + CP24Time2a
	// COT: 3, 5
	MMeTa1 TypeID = 0xa // 10
	// MMeNb1 indicates measured value, scaled value.
	// InformationElementType: SVA + QDS
	// COT: 2, 3, 5, 11, 12, 20, 20+G
	MMeNb1 TypeID = 0xb // 11
	// MMeTb1 indicates measured value, scaled value with time tag CP24Time2a.
	// InformationElementType: SVA + QDS + CP24Time2a
	// COT: 3, 5
	MMeTb1 TypeID = 0xc // 12
	// MMeNc1 indicates measured value, short floating point value.
	// InformationElementType: IEEE754STD + QDS
	// COT: 2, 3, 5, 11, 12, 20, 20+G
	MMeNc1 TypeID = 0xd // 13
	// MMeTc1 indicates measured value, short floating point value with time tag CP24Time2a.
	// InformationElementType: IEEE754STD + QDS + CP24Time2a
	// COT: 2, 3, 5, 11, 12, 20, 20+G
	MMeTc1 TypeID = 0xe // 14
	// MItNa1 indicates integrated totals.
	// InformationElementType: BCR
	// COT: 2, CotReqcogen, 37+G
	MItNa1 TypeID = 0xf // 15
	// MItTa1 indicates integrated totals with time tag CP24Time2a.
	// InformationElementType: BCR + CP24Time2a
	// COT: 3, CotReqcogen, 37+G
	MItTa1 TypeID = 0x10 // 16
	// MMeNd1 indicates measured value, normalized value without quality descriptor.
	// InformationElementType: NVA
	// COT: 1,2,3,5,11,12,20,20+G
	// [遥测 - 归一化值 - 不带时标 - 不带品质描述]
	MMeNd1 TypeID = 0x15 // 21

	// Process telegrams with long time tag (7 bytes)

	// MSpTb1 indicates single point information with time tag CP56Time2a.
	// InformationElementType: SIQ + CP56Time2a
	// COT: 3,5,11,12
	MSpTb1 TypeID = 0x1e // 30
	// MDpTb1 indicates double point information with time tag CP56Time2a.
	// InformationElementType: DIQ + CP56Time2a
	// COT: 3,5,11,12
	MDpTb1 TypeID = 0x1f // 31
	// MMeTd1 indicates measured value, normalized value with time tag CP56Time2a.
	// InformationElementType: NVA + QDS + CP56Time2a
	// COT: CotSpont, 5
	MMeTd1 TypeID = 0x22 // 34
	// MMeTe1 indicates measured value, scaled value with time tag CP56Time2a.
	// InformationElementType: SVA + QDS + CP56Time2a
	// COT: CotSpont, 5
	MMeTe1 TypeID = 0x23 // 35
	// MMeTf1 indicates measured value, short floating point value with time tag CP56Time2a.
	// InformationElementType: IEEE754STD + QDS + CP56Time2a
	// COT: 2, CotSpont, 5, 11, 12, 20, 20+G
	MMeTf1 TypeID = 0x24 // 36
	// MItTb1 indicates integrated totals with time tag CP56Time2a.
	// InformationElementType: BCR + CP56Time2a
	// COT: CotSpont, CotReqcogen, 37+G
	MItTb1 TypeID = 0x25 // 37

	// Process information in control direction.

	// CScNa1 indicates single command.
	// InformationElementType: SCO
	// COT: 6, 7, 8, 9, 10, 44, 45, 46, 47
	CScNa1 TypeID = 0x2d // 45
	// CDcNa1 indicates double command.
	// InformationElementType: DCO
	// COT: 6, 7, 8, 9, 10, 44, 45, 46, 47
	CDcNa1 TypeID = 0x2e // 46
	// CRcNa1 indicates regulating step command.
	// InformationElementType: RCO
	// COT: 6, 7, 8, 9, 10, 44, 45, 46, 47
	CRcNa1 TypeID = 0x2f // 47
	// CSeNa1 indicates set-point command, normalized value.
	// InformationElementType: NVA + QOS
	// COT: 6, 7, 8, 9, 10, 44, 45, 46, 47
	CSeNa1 TypeID = 0x30 // 48
	// CSeNb1 indicates set-point command, scaled value.
	// InformationElementType: SVA + QOS
	// COT: 6, 7, 8, 9, 10, 44, 45, 46, 47
	CSeNb1 TypeID = 0x31 // 49
	// CSeNc1 indicates set-point command, short floating point value.
	// InformationElementType: IEEE754STD + QOS
	// COT: 6, 7, 8, 9, 10, 44, 45, 46, 47
	CSeNc1 TypeID = 0x32 // 50

	// Command telegrams with long time tag.

	// CScTa1 indicates single command with time tag CP56Time2a.
	// InformationElementType: SCO + CP56Time2a
	CScTa1 TypeID = 0x3a // 58
	// CDcTa1 indicates double command with time tag CP56Time2a.
	// InformationElementType: DCO + CP56Time2a
	CDcTa1 TypeID = 0x3b // 59
	// CSeTa1 indicates set-point command, normalized value with time tag CP56Time2a.
	// InformationElementType: NVA + QOS + CP56Time2a
	CSeTa1 TypeID = 0x3d // 61
	// CSeTb1 indicates set-point command, scaled value with time tag CP56Time2a.
	// InformationElementType: SVA + QOS + CP56Time2a
	CSeTb1 TypeID = 0x3e // 62
	// CSeTc1 indicates set-point command, short floating point value with time tag CP56Time2a.
	// InformationElementType: IEEE754STD + QOS + CP56Time2a
	CSeTc1 TypeID = 0x3f // 63

	// System information in control direction.

	// CIcNa1 indicates general interrogation command. [召唤全数据]
	// InformationElementType: QOI
	// COT: 6,7,8,9,10,44,45,46,47
	CIcNa1 TypeID = 0x64 // 100
	// CCiNa1 indicates counter interrogation command. [召唤全电度]
	// InformationElementType: QCC
	// COT: 6,7,8,9,10,44,45,46,47
	CCiNa1 TypeID = 0x65 // 101
	// CRdNa1 indicates read command.
	// InformationElementType: null
	// COT: 5
	CRdNa1 TypeID = 0x66 // 102
	// CCsNa1 indicates clock synchronization command. [时钟同步]
	// InformationElementType: CP56Time2a
	// COT: 3,6,7,44,45,46,47
	CCsNa1 TypeID = 0x67 // 103
	// CTsNb1 indicates test command.
	// InformationElementType: FBP
	// COT: 6, 7, 44, 45, 46, 47
	CTsNb1 TypeID = 0x68 // 104
	// CRpNc1 indicates reset process command.
	// InformationElementType: QRP
	// COT: 6, 7, 44, 45, 46, 47
	CRpNc1 TypeID = 0x69 // 105
	// CCdNa1 indicates delay acquisition command.
	// InformationElementType: CP16Time2a
	// COT: CotAct, CotActCon, 44, 45, 46, 47
	CCdNa1 TypeID = 0x6a // 106
	// CTsTa1 indicates command with time tag CP56Time2a.
	// InformationElementType:
	CTsTa1 TypeID = 0x6b // 107
)

func (asdu *ASDU) parseTypeID(data byte) TypeID {
	asdu.typeID = TypeID(data)
	return asdu.typeID
}

/*
SQ (Structure Qualifier, 1 bit) specifies how information objects or elements are addressed.
  - SQ=0 (false): each ASDU contains one or more than one equal information objects:
    | <-              8 bits              -> |
    | Type Identification              [1B]  | --------------------
    | 0 | Number of objects            [7b]  |           |
    | T | P/N | Cause of transmission  [6b]  | Data Unit Identifier
    | Original address (ORG)           [1B]  |           |
    | ASDU address fields              [2B]  | --------------------
    | Information object address (IOA) [3B]  | --------------------
    | Information Elements                   | Information Object 1
    | Time Tag (if used)                     | --------------------
    | Information object address (IOA) [3B]  | --------------------
    | Information Elements                   | Information Object 2
    | Time Tag (if used)                     | --------------------
    | Information object address (IOA) [3B]  | --------------------
    | Information Elements                   | Information Object N
    | Time Tag (if used)                     | --------------------
    | <-              SQ = 0              -> |
  - the number of objects is binary coded (NumberOfObjects), and defines the number of the information objects;
  - each information object has its own information object address (IOA);
  - each single element or a combination of elements of object is addressed by the IOA.
  - [personal guess] SQ=0 is used to transmit a set of discontinuous values.
  - SQ=1  (true): each ASDU contains just one information object.
    | <-              8 bits              -> |
    | Type Identification              [1B]  | --------------------
    | 1 | Number of objects            [7b]  |           |
    | T | P/N | Cause of transmission  [6b]  | Data Unit Identifier
    | Original address (ORG)           [1B]  |           |
    | ASDU address fields              [2B]  | --------------------
    | Information object address (IOA) [3B]  | --------------------
    | Information Element 1                  |           |
    | Information Element 2                  |           |
    | Information Element 3                  | Information Object
    | Information Element N                  |           |
    | Time Tag (if used)                     | --------------------
    | <-              SQ = 1              -> |
  - the number of elements is binary coded (NumberOfObjects), and defines the number of the information elements;
  - there is just one information object address, which is the address of the first information element, the following
    information elements are identified by numbers continuous by +1 from this offset;
  - all information elements are of the same format, such as a measured value.
  - [personal guess] SQ=1 is used to transmit a sequence of continuous values to save bandwidth.
*/
type SQ bool

func (asdu *ASDU) parseSQ(data byte) SQ {
	asdu.sq = (data & (1 << 7)) == 1<<7
	return asdu.sq
}

/*
NOO (Number of Objects/Elements, 7 bits).
*/
type NOO = uint8

func (asdu *ASDU) parseNOO(data byte) NOO {
	asdu.nObjs = data & 0b1111111
	return asdu.nObjs
}

/*
T (Test, 1 bit) defines ASDUs which generated during test conditions. That is to say, it is not intended to control the
process or change the system state.
- T=0 (false): no test, used in the product environment.
- T=1  (true): test, used in the development environment.
*/
type T bool // Test

func (asdu *ASDU) parseT(data byte) T {
	asdu.t = (data & (1 << 7)) == 1<<7
	return asdu.t
}

/*
PN (Positive/Negative, 1 bit) indicates the positive or negative confirmation of an activation requested by a primary
application function. The bit is used when the control command is mirrored in the monitor direction, and it provides
indication of whether the command was executed or not.
- PN=0 (false): positive confirm.
- PN=1  (true): negative confirm.
*/
type PN bool

func (asdu *ASDU) parsePN(data byte) PN {
	asdu.pn = (data & (1 << 6)) == 1<<6
	return asdu.pn
}

/*
COT (Cause of Transmission, 6 bits) is used to control message routing.
- value range:

  - 0 is not defined!

  - 1-47 is used for standard IEC 101 definitions

  - 48-63 is for special use (private range)

  - COT field is used to control the routing of messages both on the communication network, and within a station,
    directing by ASDU to the correct program or task for processing. ASDUs in control direction are confirmed application
    services and may be mirrored in monitor direction with different causes of transmission.

  - COT is a 6-bit code which is used in interpreting the information at the destination station. Each defined ASDU
    type has a defined subset of the codes which are meaningful with it.
*/
type COT uint8

const (
	// the standard definitions of COT

	// 14-19 is reserved for further compatible definitions

	CotPerCyc               COT = 1  // periodic, cyclic
	CotBack                 COT = 2  // background scan
	CotSpont                COT = 3  // spontaneous
	CotInit                 COT = 4  // initialized
	CotReq                  COT = 5  // request or requested
	CotAct                  COT = 6  // activation
	CotActCon               COT = 7  // activation confirmation
	CotDeact                COT = 8  // deactivation
	CotDeactCon             COT = 9  // deactivation confirmation
	CotActTerm              COT = 10 // activation termination
	CotRetRem               COT = 11 // return information caused by a remote command
	CotRetLoc               COT = 12 // return information caused by a local command
	CotFile                 COT = 13 // file transfer
	CotInrogen              COT = 20 // interrogated by general interrogation
	CotInro1                COT = 21 // interrogated by general interrogation group1
	CotInro2                COT = 22 // interrogated by general interrogation group2
	CotInro3                COT = 23 // interrogated by general interrogation group3
	CotInro4                COT = 24 // interrogated by general interrogation group4
	CotInro5                COT = 25 // interrogated by general interrogation group5
	CotInro6                COT = 26 // interrogated by general interrogation group6
	CotInro7                COT = 27 // interrogated by general interrogation group7
	CotInro8                COT = 28 // interrogated by general interrogation group8
	CotInro9                COT = 29 // interrogated by general interrogation group9
	CotInro10               COT = 30 // interrogated by general interrogation group10
	CotInro11               COT = 31 // interrogated by general interrogation group11
	CotInro12               COT = 32 // interrogated by general interrogation group12
	CotInro13               COT = 33 // interrogated by general interrogation group13
	CotInro14               COT = 34 // interrogated by general interrogation group14
	CotInro15               COT = 35 // interrogated by general interrogation group15
	CotInro16               COT = 36 // interrogated by general interrogation group16
	CotReqcogen             COT = 37 // interrogated by counter interrogation
	CotReqco1               COT = 38 // interrogated by counter interrogation group 1
	CotReqco2               COT = 39 // interrogated by counter interrogation group 2
	CotReqco3               COT = 40 // interrogated by counter interrogation group 3
	CotReqco4               COT = 41 // interrogated by counter interrogation group 4
	CotUnknownType          COT = 44 // type identification unknown
	CotUnknownCause         COT = 45 // cause of transmission unknown
	CotUnknownAsduAddress   COT = 46 // ASDU address unknown
	CotUnknownObjectAddress COT = 47 // information object address unknown

	// TODO How to support COT for special use?
)

func (asdu *ASDU) parseCOT(data byte) COT {
	asdu.cot = COT(data & 0b111111)
	return asdu.cot
}

/*
ORG (Originator Address, 1 byte) provides a method for a controlling station to explicitly identify itself.
  - The originator address is optional when there is only one controlling station in a system. If it is not used, all bits
    are set to zero.
  - It is required when where is more than one controlling station, or some stations are dual-mode. In this case,
    the address can be used to direct command confirmations back to the particular controlling station rather than to the
    whole system.
  - If there is more than one single source in a system defined, the ASDUs in monitor direction have to be directed to
    all relevant sources of the system. In this case the specific affected source has to select its specific ASDUs.

TODO What's the differences between ORG and TCP endpoint (IP + PORT)? Can we identify the source by TCP endpoint?
*/
type ORG uint8

func (asdu *ASDU) parseORG(data byte) ORG {
	asdu.org = ORG(data)
	return asdu.org
}

/*
COA (Common Address of ASDU, 2 bytes) is normally interpreted as a station address.
- COA is either 1 or 2 bytes in length, fixed on pre-system basis. The value range of 2 bytes (the standard):
  - 0 is not used;
  - 1-65534 means a station address;
  - 65535 means global address, and it is broadcast in control direction have to be answered in monitor direction by
    the address that is the specific defined common address (station address).
  - Global Address is used when the same application function must be initiated simultaneously. It's restricted to the
    following ASDUs:
  - TypeID = CIcNa1: replay with particular system data snapshot at common time
  - TypeID = CCiNa1: freeze totals at common time
  - TypeID = CCsNa1: synchronize clocks to common time
  - TypeID = CRpNc1: simultaneous reset
*/
type COA = uint16

func (asdu *ASDU) parseCOA(data []byte) COA {
	asdu.coa = binary.LittleEndian.Uint16([]byte{data[0], data[1]})
	return asdu.coa
}
