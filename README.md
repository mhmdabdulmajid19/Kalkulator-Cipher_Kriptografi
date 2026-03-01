# 🔐 Kriptografi Klasik — Web App (Go + HTML/JS)
Tugas Semester Genap 2025/2026 · Departemen Teknik Komputer · UNDIP

## Struktur Project

```
kripto/
├── main.go              ← HTTP server + routing API
├── go.mod               ← Go module
├── cipher/
│   └── cipher.go        ← Implementasi 5 algoritma cipher
└── static/
    └── index.html       ← Frontend (HTML/CSS/JS)
```

## Cara Menjalankan

### 1. Pastikan Go sudah terinstall
```bash
go version   # minimal Go 1.21
```
Download di: https://go.dev/dl/

### 2. Jalankan server
```bash
cd kripto
go run main.go
```

### 3. Buka browser
```
http://localhost:8080
```

## API Endpoints

| Method | Endpoint        | Fungsi         |
|--------|-----------------|----------------|
| POST   | /api/vigenere   | Vigenere Cipher|
| POST   | /api/affine     | Affine Cipher  |
| POST   | /api/playfair   | Playfair Cipher|
| POST   | /api/hill       | Hill Cipher    |
| POST   | /api/enigma     | Enigma Cipher  |

### Contoh Request (Vigenere)
```json
POST /api/vigenere
{
  "text": "THISPLAINTEXT",
  "key": "SONY",
  "mode": "encrypt"
}
```

### Contoh Response
```json
{
  "ok": true,
  "output": "LVVQHZNGFHRVL",
  "steps": [
    "T(19) + S(18) = 11 → L",
    "H(7) + O(14) = 21 → V",
    ...
  ]
}
```

## Algoritma yang Diimplementasikan

1. **Vigenere Cipher** — C = (P + K) mod 26
2. **Affine Cipher** — C = (aP + b) mod 26
3. **Playfair Cipher** — Enkripsi bigram dengan matriks 5×5
4. **Hill Cipher** — C = K × P (mod 26), matriks 2×2 dan 3×3
5. **Enigma/Rotor Cipher** — Simulasi 3 rotor + reflektor
