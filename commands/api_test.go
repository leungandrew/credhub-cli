package commands_test

import (
	"net/http"

	"os"

	"fmt"

	"github.com/cloudfoundry-incubator/credhub-cli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = FDescribe("API", func() {

	ItBehavesLikeHelp("api", "a", func(session *Session) {
		Expect(session.Err).To(Say("api"))
		Expect(session.Err).To(Say("SERVER"))
	})

	It("revokes existing auth tokens when setting a new api successfully with a different auth server", func() {
		newAuthServer := NewServer()

		apiServer := NewServer()
		apiServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest("GET", "/info"),
				RespondWith(http.StatusOK, `{
					"app":{"version":"0.1.0 build DEV","name":"CredHub"},
					"auth-server":{"url":"`+newAuthServer.URL()+`"}
					}`),
			),
		)

		authServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest("DELETE", "/oauth/token/revoke/5b9c9fd51ba14838ac2e6b222d487106-r"),
				RespondWith(http.StatusOK, ""),
			),
		)

		cfg := config.ReadConfig()
		cfg.AuthURL = authServer.URL()
		cfg.AccessToken = "fake_token"
		cfg.RefreshToken = "5b9c9fd51ba14838ac2e6b222d487106-r"
		config.WriteConfig(cfg)

		session := runCommand("api", apiServer.URL())
		newCfg := config.ReadConfig()

		Eventually(session).Should(Exit(0))
		Expect(authServer.ReceivedRequests()).Should(HaveLen(1))
		Expect(newCfg.AccessToken).To(Equal("revoked"))
		Expect(newCfg.RefreshToken).To(Equal("revoked"))
	})

	It("leaves existing auth tokens intact when setting a new api with the same auth server", func() {
		apiServer := NewServer()
		apiServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest("GET", "/info"),
				RespondWith(http.StatusOK, `{
					"app":{"version":"my-version","name":"CredHub"},
					"auth-server":{"url":"`+authServer.URL()+`"}
					}`),
			),
		)

		cfg := config.ReadConfig()
		cfg.AccessToken = "fake_token"
		cfg.RefreshToken = "fake_refresh"
		config.WriteConfig(cfg)

		session := runCommand("api", apiServer.URL())

		Eventually(session).Should(Exit(0))
		newCfg := config.ReadConfig()
		Expect(newCfg.AccessToken).To(Equal("fake_token"))
		Expect(newCfg.RefreshToken).To(Equal("fake_refresh"))
		Expect(authServer.ReceivedRequests()).Should(HaveLen(0))
	})

	It("retains existing tokens when setting the api fails", func() {
		apiServer := NewServer()
		apiServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest("GET", "/info"),
				RespondWith(http.StatusNotFound, ""),
			),
		)

		cfg := config.ReadConfig()
		cfg.AuthURL = authServer.URL()
		cfg.AccessToken = "fake_token"
		cfg.RefreshToken = "fake_refresh"
		config.WriteConfig(cfg)

		session := runCommand("api", apiServer.URL())

		Eventually(session).Should(Exit(1))
		newCfg := config.ReadConfig()
		Expect(newCfg.AccessToken).To(Equal("fake_token"))
		Expect(newCfg.RefreshToken).To(Equal("fake_refresh"))
		Expect(authServer.ReceivedRequests()).Should(HaveLen(0))
	})

	FContext("when the provided server url's scheme is https", func() {
		var (
			theServer    *Server
			theServerUrl string
		)

		BeforeEach(func() {
			fmt.Println("start of inner before")
			theServer = NewServer()
			theServerUrl = setUpServer(theServer)
			fmt.Println("end of inner before")
		})

		AfterEach(func() {
			theServer.Close()
		})

		FIt("sets the target URL", func() {
			fmt.Println("start of test")
			session := runCommand("api", theServerUrl)
			fmt.Println("a")

			Eventually(session).Should(Exit(0))
			fmt.Println("b")

			session = runCommand("api")
			fmt.Println("c")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(theServerUrl))

			fmt.Println("d")
			cfg := config.ReadConfig()
			fmt.Println("e")

			Expect(cfg.AuthURL).To(Equal("https://example.com"))
			fmt.Println("f")
		})

		It("sets the target URL using a flag", func() {
			session := runCommand("api", "-s", theServerUrl)

			Eventually(session).Should(Exit(0))

			session = runCommand("api")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(theServerUrl))
		})

		It("will prefer the command's argument URL over the flag's argument", func() {
			session := runCommand("api", theServerUrl, "-s", "woooo.com")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(theServerUrl))

			session = runCommand("api")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(theServerUrl))
		})

		Context("when the provided server url is not valid", func() {
			var (
				badServer *Server
			)

			BeforeEach(func() {
				// confirm we have original good server
				session := runCommand("api", theServerUrl)

				Eventually(session).Should(Exit(0))

				badServer = NewServer()
				badServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/info"),
						RespondWith(http.StatusNotFound, ""),
					),
				)
			})

			AfterEach(func() {
				badServer.Close()
			})

			It("retains previous target when the url is not valid", func() {
				// fail to validate on bad server
				session := runCommand("api", badServer.URL())

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("The targeted API does not appear to be valid."))

				// previous value remains
				session = runCommand("api")

				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say(theServer.URL()))
			})
		})

		Context("saving configuration from server", func() {
			It("saves config", func() {
				session := runCommand("api", theServer.URL())
				Eventually(session).Should(Exit(0))

				cfg := config.ReadConfig()
				Expect(cfg.ApiURL).To(Equal(theServer.URL()))
				Expect(cfg.AuthURL).To(Equal("https://example.com"))
				Expect(cfg.InsecureSkipVerify).To(Equal(false))
			})

			It("sets file permissions so that the configuration is readable and writeable only by the owner", func() {
				configPath := config.ConfigPath()
				os.Remove(configPath)
				session := runCommand("api", theServer.URL())
				Eventually(session).Should(Exit(0))

				statResult, _ := os.Stat(configPath)

				Expect(statResult.Mode().String(), "-rw-------")
			})

			Context("when the user skips TLS validation", func() {

				It("prints warning when --skip-tls-validation flag is present", func() {
					theServer.Close()
					theServer = NewTLSServer()
					theServerUrl = setUpServer(NewTLSServer())
					session := runCommand("api", "-s", theServerUrl, "--skip-tls-validation")

					Eventually(session).Should(Exit(0))
					Eventually(session.Out).Should(Say("Warning: The targeted TLS certificate has not been verified for this connection."))
				})

				It("sets skip-tls flag in the config file", func() {
					theServer.Close()
					theServer = NewTLSServer()
					theServerUrl = setUpServer(theServer)
					session := runCommand("api", "-s", theServerUrl, "--skip-tls-validation")

					Eventually(session).Should(Exit(0))
					cfg := config.ReadConfig()
					Expect(cfg.InsecureSkipVerify).To(Equal(true))
				})

				It("resets skip-tls flag in the config file", func() {
					cfg := config.ReadConfig()
					cfg.InsecureSkipVerify = true
					err := config.WriteConfig(cfg)
					Expect(err).NotTo(HaveOccurred())

					session := runCommand("api", "-s", theServerUrl)

					Eventually(session).Should(Exit(0))
					cfg = config.ReadConfig()
					Expect(cfg.InsecureSkipVerify).To(Equal(false))
				})

				It("using a TLS server without the skip-tls flag set will fail on certificate verification", func() {
					theServer.Close()
					theServer = NewTLSServer()
					theServerUrl = setUpServer(theServer)
					session := runCommand("api", "-s", theServerUrl)

					Eventually(session).Should(Exit(1))
					Eventually(session.Err).Should(Say("Error connecting to the targeted API"))
				})

				It("using a TLS server with the skip-tls flag set will succeed", func() {
					theServer.Close()
					theServer = NewTLSServer()
					theServerUrl = setUpServer(theServer)
					session := runCommand("api", "-s", theServerUrl, "--skip-tls-validation")

					Eventually(session).Should(Exit(0))
				})

				It("records skip-tls into config file even with http URLs (will do nothing with that value)", func() {
					session := runCommand("api", theServer.URL(), "--skip-tls-validation")
					cfg := config.ReadConfig()

					Eventually(session).Should(Exit(0))
					Expect(cfg.InsecureSkipVerify).To(Equal(true))
				})
			})
		})
	})

	XContext("when the provided server url's scheme is http", func() {
		var (
			httpServer *Server
		)

		BeforeEach(func() {
			httpServer = NewServer()

			httpServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/info"),
					RespondWith(http.StatusOK, `{
					"app":{"version":"my-version","name":"CredHub"},
					"auth-server":{"url":"https://example.com"}
					}`),
				),
			)
		})

		AfterEach(func() {
			httpServer.Close()
		})

		It("does not use TLS", func() {
			session := runCommand("api", httpServer.URL())
			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(httpServer.URL()))

			session = runCommand("api")

			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say(httpServer.URL()))
		})

		It("prints warning text", func() {
			session := runCommand("api", httpServer.URL())
			Eventually(session).Should(Exit(0))
			Eventually(session).Should(Say("Warning: Insecure HTTP API detected. Data sent to this API could be intercepted" +
				" in transit by third parties. Secure HTTPS API endpoints are recommended."))
		})
	})
})

func setUpServer(aServer *Server) string {
	aUrl := aServer.URL()

	aServer.AppendHandlers(
		CombineHandlers(
			VerifyRequest("GET", "/info"),
			RespondWith(http.StatusOK, `{
					"app":{"version":"0.1.0 build DEV","name":"CredHub"},
					"auth-server":{"url":"https://example.com"}
					}`),
		),
	)

	return aUrl
}
