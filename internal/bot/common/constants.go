package common

// FlowName mendefinisikan alur yang sedang aktif di sesi
type FlowName string

const (
	FlowDataDiri  FlowName = "FLOW_DATA_DIRI"
	FlowSurat     FlowName = "FLOW_SURAT"
	FlowPengaduan FlowName = "FLOW_PENGADUAN"
	FlowUlasan    FlowName = "FLOW_ULASAN"
)
