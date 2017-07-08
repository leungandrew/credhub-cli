package models_test

import (
	"github.com/cloudfoundry-incubator/credhub-cli/models"
	"github.com/mitchellh/mapstructure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// It seems silly to duplicate these tests since we know the code paths are the same
// can we delete most of one test, leaving behind just enough to show that it's still basically working
// and trust that if we ever make these use different code paths (super unlikely) that our future
// teammates will do the right thing? -- Phil
// P.S. this would all be much happier if it were seven different tests instead of one giant test, I think
var _ = Describe("CredentialBulkImport", func() {
	Describe("readBytes()", func() {
		It("parses YAML", func() {
			var credentialBulkImport models.CredentialBulkImport
			err := credentialBulkImport.ReadBytes(
				[]byte(
					`credentials:
- name: /director/deployment/blobstore - agent
  type: password
  value: gx4ll8193j5rw0wljgqo
- name: /director/deployment/blobstore - director
  type: value
  value: y14ck84ef51dnchgk4kp
- name: /director/deployment/bosh-ca
  type: certificate
  value:
    ca: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    certificate: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
- name: /director/deployment/bosh-cert
  type: certificate
  value:
    ca_name: /dan-cert
    certificate: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
- name: /director/deployment/rsa
  type: rsa
  value:
    public_key: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
- name: /director/deployment/ssh
  type: ssh
  value:
    public_key: ssh-rsa AAAAB3NzaC...C1X7
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
- name: /director/deployment/user
  type: user
  value:
    username: covfefe
    password: jidiofj1239i1293i12n3
- name: /director/deployment/json
  type: json
  value:
    arbitrary_object:
      nested_array:
      - array_val1
      - array_object_subvalue: covfefe
      - [sub_array_val1, sub_array_val2]`))

			Expect(err).To(BeNil())
			Expect(len(credentialBulkImport.Credentials)).To(Equal(8))
			Expect(credentialBulkImport.Credentials[0].Name).To(Equal("/director/deployment/blobstore - agent"))
			Expect(credentialBulkImport.Credentials[1].Name).To(Equal("/director/deployment/blobstore - director"))
			Expect(credentialBulkImport.Credentials[2].Name).To(Equal("/director/deployment/bosh-ca"))
			Expect(credentialBulkImport.Credentials[3].Name).To(Equal("/director/deployment/bosh-cert"))
			Expect(credentialBulkImport.Credentials[4].Name).To(Equal("/director/deployment/rsa"))
			Expect(credentialBulkImport.Credentials[5].Name).To(Equal("/director/deployment/ssh"))
			Expect(credentialBulkImport.Credentials[6].Name).To(Equal("/director/deployment/user"))
			Expect(credentialBulkImport.Credentials[0].Type).To(Equal("password"))
			Expect(credentialBulkImport.Credentials[1].Type).To(Equal("value"))
			Expect(credentialBulkImport.Credentials[2].Type).To(Equal("certificate"))
			Expect(credentialBulkImport.Credentials[3].Type).To(Equal("certificate"))
			Expect(credentialBulkImport.Credentials[4].Type).To(Equal("rsa"))
			Expect(credentialBulkImport.Credentials[5].Type).To(Equal("ssh"))
			Expect(credentialBulkImport.Credentials[6].Type).To(Equal("user"))
			Expect(credentialBulkImport.Credentials[7].Type).To(Equal("json"))

			Expect(credentialBulkImport.Credentials[0].Value.(string)).To(Equal("gx4ll8193j5rw0wljgqo"))
			Expect(credentialBulkImport.Credentials[1].Value.(string)).To(Equal("y14ck84ef51dnchgk4kp"))

			var certificate1 models.Certificate
			err = mapstructure.Decode(credentialBulkImport.Credentials[2].Value, &certificate1)
			Expect(err).To(BeNil())
			Expect(certificate1.Ca).To(ContainSubstring(`-----BEGIN CERTIFICATE-----`))
			Expect(certificate1.Certificate).To(ContainSubstring(`-----BEGIN CERTIFICATE-----`))
			Expect(certificate1.PrivateKey).To(ContainSubstring(`-----BEGIN RSA PRIVATE KEY-----`))
			Expect(certificate1.CaName).To(Equal(""))

			var certificate2 models.Certificate
			err = mapstructure.Decode(credentialBulkImport.Credentials[3].Value, &certificate2)
			Expect(err).To(BeNil())
			Expect(certificate2.Ca).To(Equal(""))
			Expect(certificate2.Certificate).To(ContainSubstring(`-----BEGIN CERTIFICATE-----`))
			Expect(certificate2.PrivateKey).To(ContainSubstring(`-----BEGIN RSA PRIVATE KEY-----`))
			Expect(certificate2.CaName).To(Equal("/dan-cert"))

			var rsa models.RsaSsh
			err = mapstructure.Decode(credentialBulkImport.Credentials[4].Value, &rsa)
			Expect(err).To(BeNil())
			Expect(rsa.PublicKey).To(ContainSubstring(`-----BEGIN CERTIFICATE-----`))
			Expect(rsa.PrivateKey).To(ContainSubstring(`-----BEGIN RSA PRIVATE KEY-----`))

			var ssh models.RsaSsh
			err = mapstructure.Decode(credentialBulkImport.Credentials[5].Value, &ssh)
			Expect(err).To(BeNil())
			Expect(ssh.PublicKey).To(ContainSubstring(`ssh-rsa AAAAB3NzaC...C1X7`))
			Expect(ssh.PrivateKey).To(ContainSubstring(`-----BEGIN RSA PRIVATE KEY-----`))

			var user models.User
			err = mapstructure.Decode(credentialBulkImport.Credentials[6].Value, &user)
			Expect(err).To(BeNil())
			Expect(user.Username).To(ContainSubstring(`covfefe`))
			Expect(user.Password).To(ContainSubstring(`jidiofj1239i1293i12n3`))

			jsonAsMap := credentialBulkImport.Credentials[7].Value.(map[string]interface{})
			Expect(jsonAsMap["arbitrary_object"]).NotTo(BeNil())
			arbitraryObject := jsonAsMap["arbitrary_object"].(map[string]interface{})
			nestedArray := arbitraryObject["nested_array"].([]interface{})
			Expect(nestedArray).NotTo(BeNil())
			Expect(len(nestedArray)).To(Equal(3))
			Expect(nestedArray[0]).To(Equal("array_val1"))
			secondArrayValue := nestedArray[1].(map[string]interface{})
			Expect(secondArrayValue["array_object_subvalue"]).To(Equal("covfefe"))
			arrayArray := nestedArray[2].([]interface{})
			Expect(arrayArray[0].(string)).To(Equal("sub_array_val1"))
			Expect(arrayArray[1].(string)).To(Equal("sub_array_val2"))
		})
	})

	Describe("readFile()", func() {
		It("parses YAML from an input file", func() {
			var credentialBulkImport models.CredentialBulkImport
			err := credentialBulkImport.ReadFile("../test/test_import_file.yml")

			Expect(err).To(BeNil())
			Expect(len(credentialBulkImport.Credentials)).To(Equal(8))
			Expect(credentialBulkImport.Credentials[0].Name).To(Equal("/director/deployment/blobstore - agent"))
			Expect(credentialBulkImport.Credentials[1].Name).To(Equal("/director/deployment/blobstore - director"))
			Expect(credentialBulkImport.Credentials[2].Name).To(Equal("/director/deployment/bosh-ca"))
			Expect(credentialBulkImport.Credentials[3].Name).To(Equal("/director/deployment/bosh-cert"))
			Expect(credentialBulkImport.Credentials[4].Name).To(Equal("/director/deployment/rsa"))
			Expect(credentialBulkImport.Credentials[5].Name).To(Equal("/director/deployment/ssh"))
			Expect(credentialBulkImport.Credentials[6].Name).To(Equal("/director/deployment/user"))
			Expect(credentialBulkImport.Credentials[0].Type).To(Equal("password"))
			Expect(credentialBulkImport.Credentials[1].Type).To(Equal("value"))
			Expect(credentialBulkImport.Credentials[2].Type).To(Equal("certificate"))
			Expect(credentialBulkImport.Credentials[3].Type).To(Equal("certificate"))
			Expect(credentialBulkImport.Credentials[4].Type).To(Equal("rsa"))
			Expect(credentialBulkImport.Credentials[5].Type).To(Equal("ssh"))
			Expect(credentialBulkImport.Credentials[6].Type).To(Equal("user"))

			Expect(credentialBulkImport.Credentials[0].Value.(string)).To(Equal("gx4ll8193j5rw0wljgqo"))
			Expect(credentialBulkImport.Credentials[1].Value.(string)).To(Equal("y14ck84ef51dnchgk4kp"))

			var certificate1 models.Certificate
			err = mapstructure.Decode(credentialBulkImport.Credentials[2].Value, &certificate1)
			Expect(err).To(BeNil())
			Expect(certificate1.Ca).To(ContainSubstring(`ca-certificate`))
			Expect(certificate1.Certificate).To(ContainSubstring(`certificate`))
			Expect(certificate1.PrivateKey).To(ContainSubstring(`private-key`))
			Expect(certificate1.CaName).To(Equal(""))

			var certificate2 models.Certificate
			err = mapstructure.Decode(credentialBulkImport.Credentials[3].Value, &certificate2)
			Expect(err).To(BeNil())
			Expect(certificate2.Ca).To(Equal(""))
			Expect(certificate2.Certificate).To(ContainSubstring(`certificate`))
			Expect(certificate2.PrivateKey).To(ContainSubstring(`private-key`))
			Expect(certificate2.CaName).To(Equal("/dan-cert"))

			var rsa models.RsaSsh
			err = mapstructure.Decode(credentialBulkImport.Credentials[4].Value, &rsa)
			Expect(err).To(BeNil())
			Expect(rsa.PublicKey).To(ContainSubstring(`public-key`))
			Expect(rsa.PrivateKey).To(ContainSubstring(`private-key`))

			var ssh models.RsaSsh
			err = mapstructure.Decode(credentialBulkImport.Credentials[5].Value, &ssh)
			Expect(err).To(BeNil())
			Expect(ssh.PublicKey).To(ContainSubstring(`ssh-public-key`))
			Expect(ssh.PrivateKey).To(ContainSubstring(`private-key`))

			var user models.User
			err = mapstructure.Decode(credentialBulkImport.Credentials[6].Value, &user)
			Expect(err).To(BeNil())
			Expect(user.Username).To(ContainSubstring(`covfefe`))
			Expect(user.Password).To(ContainSubstring(`jidiofj1239i1293i12n3`))

			jsonAsMap := credentialBulkImport.Credentials[7].Value.(map[string]interface{})
			Expect(jsonAsMap["arbitrary_object"]).NotTo(BeNil())
			arbitraryObject := jsonAsMap["arbitrary_object"].(map[string]interface{})
			nestedArray := arbitraryObject["nested_array"].([]interface{})
			Expect(nestedArray).NotTo(BeNil())
			Expect(len(nestedArray)).To(Equal(2))
			Expect(nestedArray[0]).To(Equal("array_val1"))
			secondArrayValue := nestedArray[1].(map[string]interface{})
			Expect(secondArrayValue["array_object_subvalue"]).To(Equal("covfefe"))
		})
	})
})
