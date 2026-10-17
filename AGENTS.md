# Aturan Agent

## 1. Konfirmasi Ambiguitas
- Setelah menerima prompt dari user, tanyakan dan konfirmasi kembali sebelum mengeksekusi jika ada bagian yang ambigu, tidak lengkap, atau kurang jelas.
- Tujuan: mengurangi halusinasi dan asumsi yang salah.
- Jangan langsung menebak intent user bila ada lebih dari satu interpretasi yang masuk akal — tanyakan dulu.

## 2. Audit Browser Menggunakan Brave
- Gunakan **Brave Browser** yang terinstall di sistem Manjaro Linux ini (`/usr/bin/brave`, Brave Browser 155.1.97.56) untuk semua kebutuhan audit browser.
- Contoh penggunaan: Lighthouse audit, monitoring performa, screenshot headless, dsb.
- Jalankan Lighthouse via Brave (headless) dengan pola berikut:
  ```bash
  npx lighthouse <url> --chrome-flags="--headless=new --no-sandbox" --output=html --output-path=./lighthouse-report.html
  ```
  Brave berbasis Chromium, jadi kompatibel dengan tooling Lighthouse/Chrome DevTools.
- Untuk kebutuhan browser automation lain (Puppeteer/Playwright), arahkan ke binary Brave:
  ```bash
  executablePath: '/usr/bin/brave'
  ```
- Jangan gunakan Chrome/Chromium/Firefox lain untuk audit — hanya Brave.

## 3. Node.js via nvm
- Selalu jalankan `nvm use 24.20.0` **sebelum** memakai command `npm`, `npx`, atau `node`, agar versi Node konsisten.
  ```bash
  nvm use 24.20.0 && npx <command>
  ```
- Jika `nvm` tidak ada di shell non-interaktif, source dulu: `source ~/.nvm/nvm.sh`.

## 4. Dokumentasi
- Dokumen teknis & produk ada di `.ai/` (PRD, TRD, APP-FLOW, DESIGN, IMPLEMENTATION-PLAN).
- Changelogs & deployment ada di `docs/` (`CHANGELOGS.md`, `DEPLOYMENT.md`).
- Saat menambah fitur: perbarui `README.md`, `.ai/TRD.md`, dan catat entri baru di `docs/CHANGELOGS.md`.
