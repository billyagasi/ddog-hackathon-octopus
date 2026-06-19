# Ringkasan Implementasi AI Incident Commander

## Apa yang Dibuat (What I Created)

Kami telah membangun arsitektur lengkap dari "AI Incident Commander" berdasarkan panduan dokumen yang ada. Sistem ini adalah platform *Autonomous Incident Response* yang beroperasi secara *Datadog-Native*.
Sistem ini menggunakan **FastAPI** sebagai *gateway* utama, **LangGraph** sebagai sistem orkestrasi *Multi-Agent*, dan **Amazon Bedrock** sebagai mesin *LLM/Reasoning*.

Komponen utama yang dibuat meliputi:
1. **Multi-Agent System**: Mencakup *Incident Lead AI*, *Infrastructure Operations AI*, *Application Intelligence AI*, dan *Service Management AI*.
2. **Integrasi Datadog MCP**: Modul yang menghubungkan *tools* yang dipakai agen (seperti `query_logs`, `query_traces`) langsung ke Datadog API.
3. **Integrasi Slack (War Room)**: Menggunakan Slack Bolt untuk otomatis membuat kanal (*channel*) saat insiden terjadi, menampilkan temuan *real-time*, dan memproses *approval* (Persetujuan/Penolakan) oleh manusia sebelum mengeksekusi tindakan (*remediation*).
4. **Auto RCA Engine & Database**: Menggunakan **PostgreSQL** dengan ekstensi **pgvector** untuk menyimpan riwayat insiden dan dokumen RCA (Root Cause Analysis) yang di-generate oleh LLM. Seluruh jejak kerja LLM juga terpantau lewat **Datadog LLM Observability**.

---

## Bagaimana Cara Kerjanya (How It Works)

1. **Deteksi & Trigger**: Ketika Datadog Watchdog mendeteksi sebuah anomali (misalnya *latency spike*), notifikasi dikirimkan melalui *webhook* ke aplikasi FastAPI.
2. **Pembuatan War Room**: Sistem secara otomatis membuat *channel* khusus atau "War Room" di Slack.
3. **Investigasi Paralel (LangGraph)**: *Incident Lead AI* menerima notifikasi, lalu menugaskan 3 agen spesialis secara paralel:
   - *Infra Ops* mengecek riwayat *deployment* dan kondisi infrastruktur.
   - *App Intelligence* menganalisis *traces*, *logs*, dan APM.
   - *Service Management* menghitung dampak bisnis, ketersediaan SLO, dan pengguna yang terdampak.
4. **Pengumpulan Bukti via MCP**: Setiap agen menggunakan kumpulan *tools* dari Datadog MCP untuk melakukan kueri tanpa perlu login ke sistem klien atau server *database* secara langsung.
5. **Rekomendasi & Approval**: *Incident Lead AI* menyatukan seluruh temuan, menyusun hipotesis sumber masalah (*Root Cause*), dan mengeluarkan rekomendasi aksi. Sistem mengirimkan pesan konfirmasi interaktif (tombol) ke Slack.
6. **Otomatisasi**: Setelah disetujui (*Human-in-the-Loop*), fungsi `execute_workflow` dijalankan untuk memicu Datadog Workflow Automation menyelesaikan insiden (contoh: me-*rollback* versi rilis).
7. **Generate RCA**: Setelah insiden usai, sistem memanggil Bedrock untuk merangkum seluruh *timeline*, bukti, dan keputusan menjadi dokumen RCA, lalu menyimpannya ke *database* `pgvector` sebagai bank pengetahuan (*Knowledge Base*).

---

## Kelebihan (Strengths)

- **Datadog-Native & Aman**: Sistem ini memusatkan intelijennya pada Datadog MCP. Agen AI tidak memiliki akses mentah ke infrastruktur (seperti Kubernetes atau AWS CLI) sehingga risiko *security* jauh lebih rendah.
- **Observabilitas Penuh**: Setiap *prompt*, *token*, dan durasi proses pikir LLM direkam oleh OpenTelemetry lalu dikirim ke **Datadog LLM Observability**.
- **Terjaga (Human-in-the-Loop)**: AI tidak diperbolehkan mengambil keputusan untuk memperbaiki sistem (terutama di taraf produksi) tanpa persetujuan (klik *Approve*) eksplisit dari pihak teknisi (*engineer*).
- **Pemecahan Masalah Komprehensif**: Arsitektur *Multi-Agent* mencegah "halusinasi" karena tiap agen hanya meneliti ranah keahliannya saja. Agen spesialis mencari bukti, sementara agen utama (Lead) hanya menyimpulkan jika buktinya kuat.
- **Berbasis Vektor (*pgvector*)**: Memudahkan pencarian riwayat RCA berbasis kesamaan semantik. Jika insiden serupa pernah terjadi, platform ini dapat mempelajarinya kembali dari *database*.

---

## Kekurangan (Weaknesses)

- **Ketergantungan Kuat pada Datadog (Vendor Lock-in)**: Solusi ini tidak akan bekerja maksimal jika sistem observabilitas utamanya dipindahkan ke vendor lain (misalnya Grafana atau New Relic). Solusi dibangun sepenuhnya untuk kapabilitas Datadog.
- **Biaya Konsumsi LLM (AWS Bedrock)**: Menjalankan banyak agen (yang bekerja paralel mengevaluasi dokumen JSON/logs yang besar) bisa menghabiskan *token* dengan sangat cepat. Insiden yang kompleks berpotensi memakan biaya *inference* yang cukup mahal per kasusnya.
- **Kompleksitas Manajemen State (LangGraph)**: Memelihara proses *asynchronous* yang menunggu masukan dari manusia (*approval*) dengan sistem arsitektur grafik *(Graph)* bisa rentan mengalami status gantung (*stuck*) apabila instruksi ke agen atau respons integrasi (seperti Slack) gagal di tengah jalan.
- **Tergantung Kecepatan API**: Cepat atau lambatnya AI menangani insiden akan sangat dipengaruhi oleh limit beban (Rate Limiting) dan *latency* dari sisi respons Datadog API itu sendiri, khususnya ketika me-kueri data log yang masif.
