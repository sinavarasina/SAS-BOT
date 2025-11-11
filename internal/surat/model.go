package surat

type JenisSurat string

const (
	DOMISILI        JenisSurat = "sk_domisili.tex"
	USAHA           JenisSurat = "sk_usaha.tex"
	SKTM_UMUM       JenisSurat = "sktm_umum.tex"
	SKTM_TANGGUNGAN JenisSurat = "sktm_tanggungan.tex"
	KEMATIAN        JenisSurat = "sk_kematian.tex"
)

//
var BaseFields = []string{
	"NAMA",
	"TTL",
	"JK",
	"AGAMA",
	"NIK",
	"PEKERJAAN",
	"ALAMAT",
}
// }

// i think to reduce complexity, i just make it specific.

var SuratFields = map[JenisSurat][]string{
	DOMISILI: {"ALASANPERLU"}, // Hanya tanya 1 field
	USAHA:    {"TTLnU", "ALAMATDOM", "DUSUN"}, // Hanya tanya 3 field
	SKTM_UMUM: {"TTLnU", "ALASANPERLU"}, // Hanya tanya 2 field
	
	// SKTM Tanggungan & Kematian tidak pakai data NIK, jadi tanya semua.
	SKTM_TANGGUNGAN: {
		"NAMA.P", "JK.P", "TTL.P", "NIK.P", "PEKERJAAN.P", "AGAMA.P", "STATUS.P", "ALAMAT.P",
		"NAMA.C", "JK.C", "TTL.C", "NIK.C", "PEKERJAAN.C", "AGAMA.C", "ALAMAT.C",
		"ALASANPERLU",
	},
	KEMATIAN: {
		"FNAMAnAL", "BINoBINTI", "TTLnU", "AGAMA", "PEKERJAAN", "POSISITERAKHIR",
		"HARI", "TGL", "JAM", "TEMPAT", "ALASANMENINGGAL",
	},
}
