# 🤖 SAS-BOT

**SAS - Sindang Anom Service Bot**

Bot WhatsApp untuk membantu kebutuhan administrasi dan layanan publik Desa Sindang Anom, Kecamatan Sekampung Udik, Kabupaten Lampung Timur.

---

## 📋 Daftar Isi

- [Fitur Utama](#fitur-utama)
- [Tech Stack](#tech-stack)
- [Struktur Proyek](#struktur-proyek)
- [Instalasi](#instalasi)
- [Konfigurasi](#konfigurasi)
- [Penggunaan](#penggunaan)
- [Alur Fitur](#alur-fitur)
- [API & Integrasi](#api--integrasi)

---

## ✨ Fitur Utama

### 1. **Menu Data Diri** (Menu 1)

Penduduk dapat mendaftarkan dan mengelola data kependudukan mereka dengan lengkap.

**Sub-fitur:**

- **Input Data Diri (Baru)** - Pendaftaran data pribadi 39 field komprehensif
- **Ubah Data Diri (Pakai NIK)** - Edit data yang sudah terdaftar
- **Validasi NIK** - Pengecekan duplikasi NIK otomatis
- **Sinkronisasi Google Sheets** - Data tersimpan di database lokal dan cloud
- **Ulasan Layanan** - Rating 1-5 untuk Input Data Diri

**Data yang dikumpulkan (39 field):**

- **Identitas:** Dusun, RT, No. KK, NIK, Nama, Jenis Kelamin
- **Pribadi:** Tempat Lahir, Tanggal Lahir, Agama, Suku, Warganegara
- **Pendidikan:** Pendidikan dalam KK, Pendidikan Sedang Ditempuh
- **Pekerjaan & Status:** Pekerjaan, Status Kawin, Status Hubungan dalam KK, Status Dasar
- **Keluarga:** Nama Ayah, NIK Ayah, Nama Ibu, NIK Ibu
- **Kesehatan:** Golongan Darah, Cacat, Status Hamil, Cara KB
- **Dokumen:** Akta Lahir, Akta Perkawinan, Akta Perceraian, Paspor, KITAS
- **Identifikasi:** KTP Elektronik, Status Rekam, Tag Card
- **Kesejahteraan:** Asuransi, No. Asuransi, Alamat Sekarang

---

### 2. **Menu Pengajuan Surat** (Menu 2)

Penduduk dapat membuat permintaan surat desa secara online dengan validasi otomatis.

**Jenis Surat yang Tersedia:**

1. **Surat Domisili** - Bukti alamat tinggal
2. **Surat Usaha (SKU)** - Pendukung perizinan usaha
3. **SKTM Umum** - Surat Keterangan Tidak Mampu (umum)
4. **SKTM Tanggungan** - SKTM untuk tanggungan keluarga
5. **Surat Kematian** - Keterangan kematian penduduk

**Sub-fitur:**

- **Ajukan Surat Baru** - Pengajuan dengan NIK validation
- **Auto-fill Data** - Data otomatis terisi dari database
- **Input Field Tambahan** - Pertanyaan dinamis sesuai jenis surat
- **Konfirmasi & Edit** - Review sebelum submit
- **Generate PDF** - Surat otomatis dibuat dalam format PDF
- **Cek Progres Surat** - Tracking status surat dengan ID unik
- **Notifikasi Kades** - Admin desa diberi notifikasi untuk approval
- **Ulasan Layanan** - Rating 1-5 untuk Pengajuan Surat

---

### 3. **Menu Pengaduan** (Menu 3)

Masyarakat dapat melaporkan keluhan dan masalah dengan bukti foto.

**Sub-fitur:**

- **Ajukan Pengaduan** - Submit keluhan dengan foto & deskripsi
- **NIK Validation** - Verifikasi identitas sebelum pengajuan
- **Upload Foto** - Gambar otomatis upload ke ImgBB (cloud image hosting)
- **Deskripsi Text** - Caption untuk menjelaskan masalah
- **ID Tracking** - Pengaduan mendapat nomor unik (P-XXX)
- **Cek Status Pengaduan** - Lacak perkembangan laporan Anda
- **Ulasan Layanan** - Rating 1-5 untuk layanan Pengaduan

---

### 4. **Sistem Ulasan (Review)**

Setiap transaksi (Input Data, Pengajuan Surat, Pengaduan) memiliki tahap ulasan untuk quality feedback.

**Fitur:**

- Rating 1-5 untuk setiap layanan
- Tersimpan di Google Sheets terpisah per layanan
- Membantu evaluasi kualitas layanan

---

## 🛠 Tech Stack

### **Backend & Bot**

- **Go (Golang)** - Server dan business logic
- **whatsmeow** - WhatsApp Web API client
- **sqlx** - SQL database driver (PostgreSQL)
- **Protobuf** - Message serialization untuk WhatsApp

### **Database**

- **PostgreSQL** - Database relasional lokal
- **Google Sheets API** - Cloud spreadsheet untuk backup & reporting

### **External Services**

- **ImgBB API** - Upload & hosting gambar pengaduan
- **Google Gemini AI** - Natural Language Processing untuk respons cerdas
- **WhatsApp Web** - Platform messaging

### **File Format & Tools**

- **JSON** - Konfigurasi dan data options (dropdown menus)
- **CSV** - Penyimpanan data penduduk (datadiri.csv)
- **PDF** - Surat yang di-generate (dengan TeX/LaTeX)
- **Docker** - Containerization (opsional)

---

## 📁 Struktur Proyek
