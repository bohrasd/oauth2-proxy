package pagewriter

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Static Pages", func() {
	var customDir string
	const customRobots = "I AM A ROBOT!!!"
	var errorPage *errorPageWriter

	BeforeEach(func() {
		errorTmpl, err := template.New("").Parse("{{.Title}}")
		Expect(err).ToNot(HaveOccurred())
		errorPage = &errorPageWriter{
			template: errorTmpl,
		}

		customDir, err = ioutil.TempDir("", "oauth2-proxy-static-pages-test")
		Expect(err).ToNot(HaveOccurred())

		robotsTxtFile := filepath.Join(customDir, robotsTxtName)
		Expect(ioutil.WriteFile(robotsTxtFile, []byte(customRobots), 0600)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(customDir)).To(Succeed())
	})

	Context("Static Page Writer", func() {
		Context("With custom content", func() {
			var pageWriter *staticPageWriter

			BeforeEach(func() {
				var err error
				pageWriter, err = newStaticPageWriter(customDir, errorPage)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("WriterRobotsTxt", func() {
				It("Should write the custom robots txt", func() {
					recorder := httptest.NewRecorder()
					pageWriter.WriteRobotsTxt(recorder)

					body, err := ioutil.ReadAll(recorder.Result().Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(customRobots))

					Expect(recorder.Result().StatusCode).To(Equal(http.StatusOK))
				})
			})
		})

		Context("Without custom content", func() {
			var pageWriter *staticPageWriter

			BeforeEach(func() {
				var err error
				pageWriter, err = newStaticPageWriter("", errorPage)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("WriterRobotsTxt", func() {
				It("Should write the custom robots txt", func() {
					recorder := httptest.NewRecorder()
					pageWriter.WriteRobotsTxt(recorder)

					body, err := ioutil.ReadAll(recorder.Result().Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(string(defaultRobotsTxt)))

					Expect(recorder.Result().StatusCode).To(Equal(http.StatusOK))
				})

				It("Should serve an error if it cannot write the page", func() {
					recorder := &testBadResponseWriter{
						ResponseRecorder: httptest.NewRecorder(),
					}
					pageWriter.WriteRobotsTxt(recorder)

					body, err := ioutil.ReadAll(recorder.Result().Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(string("Internal Server Error")))

					Expect(recorder.Result().StatusCode).To(Equal(http.StatusInternalServerError))
				})
			})
		})
	})

	Context("loadStaticPages", func() {
		Context("With custom content", func() {
			Context("And a custom robots txt", func() {
				It("Loads the custom content", func() {
					pages, err := loadStaticPages(customDir)
					Expect(err).ToNot(HaveOccurred())
					Expect(pages).To(HaveLen(1))
					Expect(pages).To(HaveKeyWithValue(robotsTxtName, []byte(customRobots)))
				})
			})

			Context("And no custom robots txt", func() {
				It("returns the default content", func() {
					robotsTxtFile := filepath.Join(customDir, robotsTxtName)
					Expect(os.Remove(robotsTxtFile)).To(Succeed())

					pages, err := loadStaticPages(customDir)
					Expect(err).ToNot(HaveOccurred())
					Expect(pages).To(HaveLen(1))
					Expect(pages).To(HaveKeyWithValue(robotsTxtName, defaultRobotsTxt))
				})
			})
		})

		Context("Without custom content", func() {
			It("Loads the default content", func() {
				pages, err := loadStaticPages("")
				Expect(err).ToNot(HaveOccurred())
				Expect(pages).To(HaveLen(1))
				Expect(pages).To(HaveKeyWithValue(robotsTxtName, defaultRobotsTxt))
			})
		})
	})
})

type testBadResponseWriter struct {
	*httptest.ResponseRecorder
	firstWriteCalled bool
}

func (b *testBadResponseWriter) Write(buf []byte) (int, error) {
	if !b.firstWriteCalled {
		b.firstWriteCalled = true
		return 0, errors.New("write closed")
	}
	return b.ResponseRecorder.Write(buf)
}