package crypto

import "github.com/windzhu0514/xtool/config"

// PKCS5等同于PKCS7
type AlgoType string

const (
	AESCBCNoPadding                    AlgoType = "AES/CBC/NoPadding"
	AESCBCPKCS5Padding                 AlgoType = "AES/CBC/PKCS5Padding"
	AESCBCZeroPadding                  AlgoType = "AES/CBC/ZeroPadding"
	AESECBNoPadding                    AlgoType = "AES/ECB/NoPadding"
	AESECBPKCS5Padding                 AlgoType = "AES/ECB/PKCS5Padding"
	AESECBPZeroPaddin                  AlgoType = "AES/ECB/ZeroPadding"
	DESCBCNoPadding                    AlgoType = "DES/CBC/NoPadding"
	DESCBCPKCS5Padding                 AlgoType = "DES/CBC/PKCS5Padding"
	DESCBCZeroPadding                  AlgoType = "DES/CBC/ZeroPadding"
	DESECBNoPadding                    AlgoType = "DES/ECB/NoPadding"
	DESECBPKCS5Padding                 AlgoType = "DES/ECB/PKCS5Padding"
	DESECBZeroPadding                  AlgoType = "DES/ECB/ZeroPadding"
	TripleDESCBCNoPadding              AlgoType = "TripleDES/CBC/NoPadding"    // DESede
	TripleDESCBCPKCS5Padding           AlgoType = "TripleDES/CBC/PKCS5Padding" // DESede
	TripleDESCBCZeroPadding            AlgoType = "TripleDES/CBC/ZeroPadding"  // DESede
	TripleDESECBNoPadding              AlgoType = "TripleDES/ECB/NoPadding"    // DESede
	TripleDESECBPKCS5Padding           AlgoType = "TripleDES/ECB/PKCS5Padding" // DESede
	TripleDESECBZeroPadding            AlgoType = "TripleDES/ECB/ZeroPadding"  // DESede
	RSAECBPKCS1Padding                 AlgoType = "RSA/ECB/PKCS1Padding"
	RSAECBOAEPWithSHA1AndMGF1Padding   AlgoType = "RSA/ECB/OAEPWithSHA-1AndMGF1Padding"
	RSAECBOAEPWithSHA256AndMGF1Padding AlgoType = "RSA/ECB/OAEPWithSHA-256AndMGF1Padding"
)

func Decrypt(cfg config.Decrypt, data []byte) ([]byte, error) {
	switch AlgoType(cfg.AlgoName) {
	case AESCBCNoPadding:
	case AESCBCPKCS5Padding:
	case AESCBCZeroPadding:
	case AESECBNoPadding:
	case AESECBPKCS5Padding:
	case AESECBPZeroPaddin:
	case DESCBCNoPadding:
	case DESCBCPKCS5Padding:
	case DESCBCZeroPadding:
	case DESECBNoPadding:
	case DESECBPKCS5Padding:
	case DESECBZeroPadding:
	case TripleDESCBCNoPadding:
	case TripleDESCBCPKCS5Padding:
	case TripleDESCBCZeroPadding:
	case TripleDESECBNoPadding:
	case TripleDESECBPKCS5Padding:
	case TripleDESECBZeroPadding:
	case RSAECBPKCS1Padding:
	case RSAECBOAEPWithSHA1AndMGF1Padding:
	case RSAECBOAEPWithSHA256AndMGF1Padding:

	}

	return nil, nil
}
