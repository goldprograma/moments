package pkg

import (
	"bytes"

	"github.com/ChengjinWu/aescrypto"
	//cryptoEx "github.com/lylib/go-crypto"
)

var blockSize = 16
//var cAES = cryptoEx.NewCrypto(cryptoEx.StandardType.AES, cryptoEx.ModeType.CBC,
//	cryptoEx.PaddingType.PKCS5, cryptoEx.FormatType.HEX)

//使用PKCS7进行填充，IOS也是7
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//aes加密，填充秘钥key的16位，24,32分别对应AES-128, AES-192, or AES-256.
func AesCBCEncrypt(rawData, key []byte) ([]byte, error) {
	if len(key)%blockSize != 0 {
		key = PKCS7Padding(key, blockSize)
	}

	return aescrypto.AesCbcPkcs7Encrypt(rawData, key, key)
}

func AesCBCDncrypt(encryptData, key []byte) ([]byte, error) {
	if len(key)%blockSize != 0 {
		key = PKCS7Padding(key, blockSize)
	}

	return aescrypto.AesCbcPkcs7Decrypt(encryptData, key, key)

}
