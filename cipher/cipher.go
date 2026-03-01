package cipher

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// HELPERS 

func toAlpha(s string) string {
	s = strings.ToUpper(s)
	var b strings.Builder
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func gcd(a, b int) int {
	if b == 0 { return a }
	return gcd(b, a%b)
}

func modInverse(a, m int) (int, error) {
	for x := 1; x < m; x++ {
		if (a*x)%m == 1 { return x, nil }
	}
	return 0, errors.New("tidak ada invers modular")
}

type Result struct {
	Output string   `json:"output"`
	Steps  []string `json:"steps"`
}

//  1. VIGENERE 

func Vigenere(text, key, mode string) (*Result, error) {
	clean := toAlpha(text)
	cleanKey := toAlpha(key)
	if len(cleanKey) == 0 {
		return nil, errors.New("kunci tidak boleh kosong")
	}
	var result strings.Builder
	var steps []string
	ki := 0
	for _, ch := range clean {
		p := int(ch - 'A')
		k := int(cleanKey[ki%len(cleanKey)] - 'A')
		var c int
		if mode == "encrypt" {
			c = (p + k) % 26
			steps = append(steps, fmt.Sprintf("%c(%d) + %c(%d) = %d → %c", ch, p, cleanKey[ki%len(cleanKey)], k, c, rune('A'+c)))
		} else {
			c = ((p-k)%26 + 26) % 26
			steps = append(steps, fmt.Sprintf("%c(%d) - %c(%d) = %d → %c", ch, p, cleanKey[ki%len(cleanKey)], k, c, rune('A'+c)))
		}
		result.WriteRune(rune('A' + c))
		ki++
	}
	return &Result{Output: result.String(), Steps: steps}, nil
}

// 2. AFFINE 

func Affine(text string, a, b int, mode string) (*Result, error) {
	if gcd(a, 26) != 1 {
		return nil, fmt.Errorf("nilai a=%d tidak relatif prima dengan 26. Gunakan: 1,3,5,7,9,11,15,17,19,21,23,25", a)
	}
	clean := toAlpha(text)
	aInv, err := modInverse(a, 26)
	if err != nil { return nil, err }
	var result strings.Builder
	var steps []string
	for _, ch := range clean {
		p := int(ch - 'A')
		var c int
		if mode == "encrypt" {
			c = (a*p + b) % 26
			steps = append(steps, fmt.Sprintf("(%d×%d + %d) mod 26 = %d → %c", a, p, b, c, rune('A'+c)))
		} else {
			c = (aInv * ((p - b + 26) % 26)) % 26
			steps = append(steps, fmt.Sprintf("%d×(%d-%d) mod 26 = %d → %c", aInv, p, b, c, rune('A'+c)))
		}
		result.WriteRune(rune('A' + c))
	}
	return &Result{Output: result.String(), Steps: steps}, nil
}

//  3. PLAYFAIR 

func buildPlayfairMatrix(key string) []byte {
	const alpha = "ABCDEFGHIKLMNOPQRSTUVWXYZ"
	seen := make(map[byte]bool)
	var matrix []byte
	cleanKey := toAlpha(key)
	for i := 0; i < len(cleanKey); i++ {
		c := cleanKey[i]
		if c == 'J' { c = 'I' }
		if !seen[c] { seen[c] = true; matrix = append(matrix, c) }
	}
	for i := 0; i < len(alpha); i++ {
		c := alpha[i]
		if !seen[c] { seen[c] = true; matrix = append(matrix, c) }
	}
	return matrix
}

func findPos(matrix []byte, ch byte) (int, int) {
	if ch == 'J' { ch = 'I' }
	for i, c := range matrix {
		if c == ch { return i / 5, i % 5 }
	}
	return 0, 0
}

func Playfair(text, key, mode string) (*Result, error) {
	matrix := buildPlayfairMatrix(key)
	clean := strings.ReplaceAll(toAlpha(text), "J", "I")
	var pairs [][2]byte
	i := 0
	for i < len(clean) {
		a := clean[i]
		var b byte
		if i+1 < len(clean) { b = clean[i+1] } else { b = 'X' }
		if a == b { pairs = append(pairs, [2]byte{a, 'X'}); i++ } else { pairs = append(pairs, [2]byte{a, b}); i += 2 }
	}
	shift := 1
	if mode == "decrypt" { shift = -1 }
	var result strings.Builder
	var steps []string
	for _, pair := range pairs {
		ra, ca := findPos(matrix, pair[0])
		rb, cb := findPos(matrix, pair[1])
		var ea, eb byte
		if ra == rb {
			ea = matrix[ra*5+((ca+shift+5)%5)]; eb = matrix[rb*5+((cb+shift+5)%5)]
			steps = append(steps, fmt.Sprintf("%c%c → baris sama → %c%c", pair[0], pair[1], ea, eb))
		} else if ca == cb {
			ea = matrix[((ra+shift+5)%5)*5+ca]; eb = matrix[((rb+shift+5)%5)*5+cb]
			steps = append(steps, fmt.Sprintf("%c%c → kolom sama → %c%c", pair[0], pair[1], ea, eb))
		} else {
			ea = matrix[ra*5+cb]; eb = matrix[rb*5+ca]
			steps = append(steps, fmt.Sprintf("%c%c → persegi panjang → %c%c", pair[0], pair[1], ea, eb))
		}
		result.WriteByte(ea); result.WriteByte(eb)
	}
	return &Result{Output: result.String(), Steps: steps}, nil
}

//  4. HILL 

func matMul(A, B [][]int, mod int) [][]int {
	n := len(A); m := len(B[0]); k := len(B)
	result := make([][]int, n)
	for i := range result {
		result[i] = make([]int, m)
		for j := 0; j < m; j++ {
			for l := 0; l < k; l++ { result[i][j] += A[i][l] * B[l][j] }
			result[i][j] = ((result[i][j] % mod) + mod) % mod
		}
	}
	return result
}

func det3x3(m [][]int) int {
	return m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
		m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
		m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])
}

func matInv2x2(m [][]int, mod int) ([][]int, error) {
	det := ((m[0][0]*m[1][1]-m[0][1]*m[1][0])%mod + mod) % mod
	detInv, err := modInverse(det, mod)
	if err != nil { return nil, errors.New("matriks kunci tidak memiliki invers mod 26. Pastikan det(K) relatif prima dengan 26") }
	return [][]int{
		{((m[1][1]*detInv)%mod + mod) % mod, ((-m[0][1]*detInv)%mod + mod) % mod},
		{((-m[1][0]*detInv)%mod + mod) % mod, ((m[0][0]*detInv)%mod + mod) % mod},
	}, nil
}

func matInv3x3(m [][]int, mod int) ([][]int, error) {
	det := ((det3x3(m) % mod) + mod) % mod
	detInv, err := modInverse(det, mod)
	if err != nil {
		return nil, fmt.Errorf("det(K)=%d tidak memiliki invers mod 26. Matriks kunci tidak valid untuk dekripsi", det)
	}
	// Hitung matriks kofaktor
	cof := make([][]int, 3)
	for i := range cof { cof[i] = make([]int, 3) }
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			// Ambil minor 2x2
			var minor [][]int
			for r := 0; r < 3; r++ {
				if r == i { continue }
				var row []int
				for c := 0; c < 3; c++ {
					if c == j { continue }
					row = append(row, m[r][c])
				}
				minor = append(minor, row)
			}
			d := minor[0][0]*minor[1][1] - minor[0][1]*minor[1][0]
			sign := 1
			if (i+j)%2 != 0 { sign = -1 }
			cof[i][j] = sign * d
		}
	}
	// Adjugate = transpose kofaktor, lalu kalikan detInv mod 26
	inv := make([][]int, 3)
	for i := range inv { inv[i] = make([]int, 3) }
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			inv[i][j] = ((detInv * cof[j][i]) % mod + mod) % mod
		}
	}
	return inv, nil
}

func Hill(text string, keyMatrix [][]int, mode string) (*Result, error) {
	const alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	clean := toAlpha(text)
	n := len(keyMatrix)
	K := keyMatrix
	var steps []string
	if mode == "decrypt" {
		if n == 2 {
			inv, err := matInv2x2(keyMatrix, 26)
			if err != nil { return nil, err }
			K = inv
			steps = append(steps, fmt.Sprintf("Invers matriks 2×2: [[%d,%d],[%d,%d]]", K[0][0], K[0][1], K[1][0], K[1][1]))
		} else if n == 3 {
			inv, err := matInv3x3(keyMatrix, 26)
			if err != nil { return nil, err }
			K = inv
			steps = append(steps, fmt.Sprintf("Invers matriks 3×3: [[%d,%d,%d],[%d,%d,%d],[%d,%d,%d]]",
				K[0][0], K[0][1], K[0][2],
				K[1][0], K[1][1], K[1][2],
				K[2][0], K[2][1], K[2][2]))
		} else {
			return nil, errors.New("ukuran matriks tidak didukung")
		}
	}
	padded := clean
	for len(padded)%n != 0 { padded += "X" }
	var result strings.Builder
	for i := 0; i < len(padded); i += n {
		block := make([][]int, n)
		blockStr := ""
		for j := 0; j < n; j++ { block[j] = []int{int(padded[i+j] - 'A')}; blockStr += string(padded[i+j]) }
		out := matMul(K, block, 26)
		outStr := ""
		for _, row := range out { c := ((row[0] % 26) + 26) % 26; result.WriteByte(alpha[c]); outStr += string(alpha[c]) }
		steps = append(steps, fmt.Sprintf("Blok %q × K = %q", blockStr, outStr))
	}
	return &Result{Output: result.String(), Steps: steps}, nil
}

//  5. ROTOR CIPHER 
type RotorRequest struct {
	Text       string   `json:"text"`
	Mode       string   `json:"mode"`
	StartRotor int      `json:"start_rotor"` // posisi awal rotor (0, 1, 2, ...)
	RotorKeys  []string `json:"rotor_keys"`  // kunci substitusi tiap posisi rotor
}


func buildAlphabet(clean string) string {
	seen := make(map[byte]bool)
	var chars []byte
	for i := 0; i < len(clean); i++ {
		c := clean[i]
		if !seen[c] { seen[c] = true; chars = append(chars, c) }
	}
	// Urutkan secara abjad
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })
	return string(chars)
}


func validateRotorKey(key, alphabet string) error {
	if len(key) != len(alphabet) {
		return fmt.Errorf("panjang kunci (%d) harus sama dengan jumlah huruf unik teks (%d)", len(key), len(alphabet))
	}
	seen := make(map[rune]bool)
	for _, c := range key {
		if seen[c] { return fmt.Errorf("duplikat karakter: %c", c) }
		seen[c] = true
	}
	for _, c := range alphabet {
		if !seen[c] { return fmt.Errorf("karakter '%c' dari alphabet tidak ada di kunci", c) }
	}
	return nil
}

func RotorCipher(req RotorRequest) (*Result, error) {
	clean := toAlpha(req.Text)
	if len(clean) == 0 {
		return nil, errors.New("teks input tidak boleh kosong")
	}

	// Alphabet = huruf unik, diurutkan abjad (A,B,C,D,E bukan B,A,C,D,E)
	alphabet := buildAlphabet(clean)
	numRotors := len(req.RotorKeys)

	if numRotors == 0 {
		return nil, errors.New("minimal harus ada 1 rotor")
	}

	// Validasi semua kunci rotor
	rotorKeys := make([]string, numRotors)
	for i, k := range req.RotorKeys {
		rk := strings.ToUpper(strings.TrimSpace(k))
		if err := validateRotorKey(rk, alphabet); err != nil {
			return nil, fmt.Errorf("K%d: %v", i, err)
		}
		rotorKeys[i] = rk
	}

	// Posisi awal rotor
	pos := req.StartRotor % numRotors

	var result strings.Builder
	var steps []string

	for _, ch := range clean {
		c := byte(ch)

		// Cari index karakter di alphabet (sudah terurut abjad)
		alphaIdx := strings.IndexByte(alphabet, c)
		if alphaIdx < 0 { continue }

		var outChar byte
		var stepDesc string

		if req.Mode == "encrypt" {
			// Enkripsi: ambil karakter di index yang sama dari kunci rotor saat ini
			outChar = rotorKeys[pos][alphaIdx]
			stepDesc = fmt.Sprintf(
				"%c (idx=%d di %s) @ K%d=%s → %c",
				c, alphaIdx, alphabet, pos, rotorKeys[pos], outChar,
			)
		} else {
			// Dekripsi: cari posisi karakter di kunci rotor, lalu ambil dari alphabet
			keyIdx := strings.IndexByte(rotorKeys[pos], c)
			if keyIdx < 0 {
				return nil, fmt.Errorf("karakter '%c' tidak ditemukan di K%d", c, pos)
			}
			outChar = alphabet[keyIdx]
			stepDesc = fmt.Sprintf(
				"%c ditemukan di K%d=%s idx=%d → %c (di %s)",
				c, pos, rotorKeys[pos], keyIdx, outChar, alphabet,
			)
		}

		steps = append(steps, stepDesc)
		result.WriteByte(outChar)

		// Rotor maju ke posisi berikutnya
		pos = (pos + 1) % numRotors
	}

	return &Result{Output: result.String(), Steps: steps}, nil
}