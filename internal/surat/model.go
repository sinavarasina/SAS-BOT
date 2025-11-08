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
// var BaseFields = []string{
// 	"NAMA",
// 	"TTL",
// 	"JK",
// 	"AGAMA",
// 	"NIK",
// 	"PEKERJAAN",
// 	"ALAMAT",
// }

// i think to reduce complexity, i just make it specific.

var SuratFields = map[JenisSurat][]string{
	DOMISILI: {
		"NAMA",
		"TTL",
		"JK",
		"AGAMA",
		"NIK",
		"PEKERJAAN",
		"ALAMAT",
		"ALASANPERLU",
	},
	USAHA: {
		"NAMA",
		"JK",
		"TTLnU",
		"NIK",
		"AGAMA",
		"PEKERJAAN",
		"ALAMAT",
		"ALAMATDOM",
		"DUSUN",
	},
	SKTM_UMUM: {
		"NAMA",
		"TTLnU",
		"NIK",
		"PEKERJAAN",
		"ALAMAT",
		"ALASANPERLU",
	},
	SKTM_TANGGUNGAN: {
		"NAMA.P", // fyi it might be weird,
		"JK.P",   // but i have no idea to naming it
		"TTL.P",  // so i choose .P as Parent, .C as Child.
		"NIK.P",  // stupid but yea just change it if you dont like
		"PEKERJAAN.P",
		"AGAMA.P",
		"STATUS.P",
		"ALAMAT.P",
		"NAMA.C",
		"JK.C",
		"TTL.C",
		"NIK.C",
		"PEKERJAAN.C",
		"AGAMA.C",
		"ALAMAT.C",
		"ALASANPERLU",
	},
	KEMATIAN: {
		"FNAMAnAL",
		"BINoBINTI",
		"TTLnU",
		"AGAMA",
		"PEKERJAAN",
		"POSISITERAKHIR",
		"HARI",
		"TGL",
		"JAM",
		"TEMPAT",
		"ALASANMENINGGAL",
	},
}
