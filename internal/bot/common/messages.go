package common

import "fmt"

// GetMainMenu mengembalikan teks menu utama
func GetMainMenu() string {
	return `*SINDANG ANOM SERVICE - BOT*

Menu yang tersedia:

1. Data Diri
2. Pengajuan Surat 
3. Pengaduan 

Silakan pilih menu dengan mengetik nomor yang sesuai.`
}

func GetNikNotFoundMenu(nik string) string {
	return fmt.Sprintf("⚠️ NIK `%s` tidak terdaftar.\n\nSilakan pilih:\n1.  Ulangi Masukkan NIK (jika salah ketik)\n2.  Daftar Baru dengan NIK ini\n\nKetik '1' atau '2'.", nik)
}

// GetSubmenuDataDiri mengembalikan submenu data diri
func GetSubmenuDataDiri() string {
	return `*Menu Data Diri*

Menu ini digunakan untuk mengelola data kependudukan Anda.

1. Input Data Diri (Baru)
2. Edit Data Diri (Berdasarkan NIK)

Silakan pilih nomor atau ketik 'reset' untuk kembali ke menu utama.`
}

// GetSubmenuSurat mengembalikan submenu pengajuan surat
func GetSubmenuSurat() string {
	return `*Menu Pengajuan Surat*

1. Ajukan Surat Baru
2. Cek Progres Surat

Ketik nomor (1-2) atau 'reset' untuk batal.`
}

// GetSubmenuPengaduan mengembalikan submenu pengaduan
func GetSubmenuPengaduan() string {
	return `*Menu Pengaduan*

Menu ini berfungsi untuk mengelola data pengaduan masyarakat.

1. Ajukan pengaduan
2. Cek progres pengaduan

Silakan pilih nomor atau ketik 'reset' untuk kembali ke menu utama.`
}

// GetUlasanMessage mengembalikan pertanyaan ulasan
func GetUlasanMessage(serviceName string) string {
	return fmt.Sprintf("✅ Terima kasih! Proses untuk *%s* telah selesai.\n\n"+
		"⭐ Sebagai langkah terakhir, mohon berikan ulasan Anda (1-5) untuk seberapa informatif layanan ini:\n"+
		"(1 = Sangat Buruk, 5 = Sangat Baik)", serviceName)
}

// GetUlasanThanksMessage mengembalikan pesan terima kasih ulasan
func GetUlasanThanksMessage() string {
	return "✅ Terima kasih banyak atas ulasan Anda! Kami sangat menghargai masukan Anda.\n\n" + GetMainMenu()
}
