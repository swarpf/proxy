package swproxy

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/elazarl/goproxy"
)

const caCert = `-----BEGIN CERTIFICATE-----
MIIDhTCCAm2gAwIBAgIUKitWGh5EJZCrQIjlQlvv+Euy9P4wDQYJKoZIhvcNAQEL
BQAwUjELMAkGA1UEBhMCREUxEzARBgNVBAgMClNvbWUtU3RhdGUxDjAMBgNVBAoM
BUx5cmV4MR4wHAYJKoZIhvcNAQkBFg9hZG1pbkBseXJleC5uZXQwHhcNMjAwMzEw
MDczMzI3WhcNMjUwMzA5MDczMzI3WjBSMQswCQYDVQQGEwJERTETMBEGA1UECAwK
U29tZS1TdGF0ZTEOMAwGA1UECgwFTHlyZXgxHjAcBgkqhkiG9w0BCQEWD2FkbWlu
QGx5cmV4Lm5ldDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAN6hfIYB
Hr4r1Ydyh19XoDmZter0HbYGObp1eZsktoz8RFMVdAM7hZL2/IdN4+JD7ettNlap
ikU6YdhpGWZ/EIxsxSxRfChgeH63KFAo1wPwmWhH/1KIAWyesrVsGtBwaDZtlzP5
ABhLp0uwCOa+Y8N1l/p5VAAPo0kM03nsKB0NltxXft6lRZwQBrwpfxVDgj+UKtWx
x3PWai2ZqnRsRaZLAzznvuGaCCvMzrtd88LHpT/04zlBl0HhrxkAQl9q/sBZx9oe
/GBgddsJYoK3ypa0C4ae0O0sBDT1IcREc0GCfahSJ19BNRc2CB3ywVd1ClJ6Klp3
vpZH2tUkLOdQOUMCAwEAAaNTMFEwHQYDVR0OBBYEFGhjFHLLp24YiKi9igVeXUNo
KZ0hMB8GA1UdIwQYMBaAFGhjFHLLp24YiKi9igVeXUNoKZ0hMA8GA1UdEwEB/wQF
MAMBAf8wDQYJKoZIhvcNAQELBQADggEBAJOuVS4sastdJql7wNqm0a2oVwlViXzw
m/2g5wzhaU5bzNnUPdtL2Km132E0iSKw8R3cSmQWA4PeHx+Mx0Q24twJY6MeQOLz
v0T4KcE3G04AbsQbVpVusw2yePpjMlxtyWVysBKH2NmNgjqChQVDiYHf2a8MMfqj
8CIrtJedtFjgQxcFV3uxsX1W/SxAxk0CcVbz9Tk6qfsHxyT53QPWPW6KFW0hA0yU
j+qismg4pD2AOZy2cRsi7PF8aKBDw2ZZBl2uPseruQRizHu6jiDcx0odgB0aN4i+
u3/0S0r+lgWkrc8f3Lnvl4Rd+fDtoBH4aXrEHAE+7MIvQUL35um2mRA=
-----END CERTIFICATE-----`

const caKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEA3qF8hgEevivVh3KHX1egOZm16vQdtgY5unV5myS2jPxEUxV0
AzuFkvb8h03j4kPt6202VqmKRTph2GkZZn8QjGzFLFF8KGB4frcoUCjXA/CZaEf/
UogBbJ6ytWwa0HBoNm2XM/kAGEunS7AI5r5jw3WX+nlUAA+jSQzTeewoHQ2W3Fd+
3qVFnBAGvCl/FUOCP5Qq1bHHc9ZqLZmqdGxFpksDPOe+4ZoIK8zOu13zwselP/Tj
OUGXQeGvGQBCX2r+wFnH2h78YGB12wligrfKlrQLhp7Q7SwENPUhxERzQYJ9qFIn
X0E1FzYIHfLBV3UKUnoqWne+lkfa1SQs51A5QwIDAQABAoIBAQCHlr5qNsBsffHc
PkpoLMvuiMkcwXRe6ce64dUgQenUT8ek+knfth6R9U6zcSK7KTf7zFXtze/iXb49
uTS5EeYYQB6N8Uq2pJp+QjqRJ25cfepQcpjzwNVtO/IHQEHMdMljbLdL9fiy01Ce
biXdslK8NiBLch1QtDV0RhV+CfAcUImmnMRjLjzZTSD01zp2DbKvb2LW5ZI48EK8
/DmBj3o+JiMjoYVHtBuG9kwu38nvavic55LIy3/S2qryfBsVfc60nyDGoUYXeWq2
vrpVBb84rcFXyx0x/RTSb5tz0WheZ9M+B1w92Z6im35OQBF/bNRZquBBr9uwiaaL
58/6HkoRAoGBAO+ULXJRwxMdjlMrSJA/GYSdzOaMRGrIrNfRl3tgLPoNIlRiWGlE
4B82fAKWMazvj4mRZI5tFNSkPl3mPlTNFxmMqs3cDrswvkOaRdrxAFDj3jzJg4DD
PaQ1kO64wRdRIo4cE1fJ4MlT5q66CXpNCOUDlgdzeLNg8RSOYzKasRUvAoGBAO3j
7UdDMq1ipTGVkoW+rPszRLFsC9OkITkY3ODOfAxIo0SXcRgOwSZRjApqZAhv00U0
n6Y8zJZ/3Kup7ouLuocLccevKSX5VmwuaT3952xeNJPc7dH4yXO3eix7HtC8Orml
tEi7cyVm9qC2mG1IcRKt6PXCKuR3bk5hPY9X6IAtAoGBAIKiQnGeYYcPy6ZP6J42
udxVCv//JeMwDwcTEs1EMOIbvUdT5K9pzedXFyF18hpA+fxiGfmLQxt7f0JGJGCq
/9h/mjbrseCiAGzuNv7eAHUa+vgcTSctznO2fZOdjDQBmpzwdB+fRYGhzRwi9r4I
OTxeyzS+4ua0il/SEAbs0HgjAoGBALlGhWy1F2kWpRYjKgTkZpEWcu/D+MoS0JVJ
me20o8RlZlNrp3dXNnODm5AZIGO5xE/oFldAjw6/8rv4E4O3hcTb0vf0ohWjRf3n
f6v6bh1mmYh3zvlzzGJFie/OzEdB8nLYbbsf0yRUNs0gqUKj4vzrWb7eRM2/freo
4GsdykTZAoGBAOZIU/ryaljR3CdKkLHgwQYuHbj6HkUTEcHYCUreoWOkb4x/bPu8
49VWz++pOvw3M0+mhz1I8O2emb4Z7mcfvOd5wnZZfW3Q/mqZGQSGwvVyLMoj5Znr
s6+BFtF/Kd/s+Gtoc3ZoU30h76FmZUbrtok5ZYAquXOVriXmxFckjDTe
-----END RSA PRIVATE KEY-----`

// todo(lyrex): this should not be a hardcoded certificate but instead be generated at runetime

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
