package surat

type JenisSurat string

const (
	DOMISILI        JenisSurat = "sk_domisili.tex"
	USAHA           JenisSurat = "sk_usaha.tex"
	SKTM_UMUM       JenisSurat = "sktm_umum.tex"
	SKTM_TANGGUNGAN JenisSurat = "sktm_tanggungan.tex"
	KEMATIAN        JenisSurat = "sk_kematian.tex"
)

var BaseFields = []string{
	"NAMA",
	"TTL",
	"JK",
	"AGAMA",
	"NIK",
	"PEKERJAAN",
	"ALAMAT",
}

var SuratFields = map[JenisSurat][]string{
	DOMISILI:        {"ALASANPERLU"},
	USAHA:           {"TTLnU", "ALAMATDOM", "DUSUN"},
	SKTM_UMUM:       {"TTLnU", "ALASANPERLU"},
	SKTM_TANGGUNGAN: {"NAMA.P", "JK.P", "TTL.P", "NIK.P", "PEKERJAAN.P", "AGAMA.P", "STATUS.P", "ALAMAT.P", "NAMA.C", "JK.C", "TTL.C", "NIK.C", "PEKERJAAN.C", "AGAMA.C", "ALAMAT.C", "ALASANPERLU"},
	KEMATIAN:        {"FNAMAnAL", "BINoBINTI", "TTLnU", "AGAMA", "PERKERJAAN", "POSISITERAKHIR", "HARI", "TGL", "JAM", "TEMPAT", "ALASANMENINGGAL"},
}
