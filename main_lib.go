package main

import "fmt"

// #cgo CFLAGS: -Isrc/inc/
// #cgo LDFLAGS: -Lc:/temp/Hannes/src/lib/ -lRiposteClient
// #include <Riposte.h>
import "C"

func Connect() {
	fmt.Printf("Invoking c library...\n")
	r := C.RiposteConnectExA(nil, nil)
	fmt.Printf("Done %v\n", r)
}
func SystemInfoExA() {
	var pinfo C.struct_riposte_system_info_ex
	r := C.RiposteGetSystemInfoExA(&pinfo)
	fmt.Printf("Done %d %v\n", r, pinfo)
	//	fmt.Printf("Version: %s\n", string(C.GoBytes(&pinfo.szVersionString, 32)))
	fmt.Printf("Version: %s\n", C.GoString(&pinfo.szVersionString[0]))
	//fmt.Printf("Message Server Id: %ul\n", C.ulong((C.ulong)(pinfo.dwNodeId)))
	fmt.Printf("Message Server Id: %d\n", C.ulong((pinfo.dwNodeId)))
}

// NextNode
//unsigned long RiposteNextNode ( unsigned long *pdwGroupId, unsigned long *pdwNodeId  );
func NextNode() {
	var GroupID C.ulong
	var NodeID C.ulong
	var r C.ulong
	GroupID = 0
	GroupID = 0
	for r == 0 {
		r = C.RiposteNextNode(&GroupID, &NodeID)
		fmt.Printf("> %d %d\n", C.ulong(GroupID), C.ulong(NodeID))
	}
	//fmt.Printf("Done %d %ul\n", r, C.ulong(GroupID))

}

//unsigned long RiposteGetMarkerW ( PMARKER pMarker, unsigned long dwGroupId, wchar_t *szAttributes);
func GetMarkerW() {

}
func ErrorStringA() {
	//unsigned long sts,
	//unsigned char *szBuf,
	//unsigned long dwBufSize
}

func main() {
	Connect()
	SystemInfoExA()
	NextNode()
}
