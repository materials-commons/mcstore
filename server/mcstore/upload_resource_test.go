package mcstore

import (
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/ezhttp"
	c "github.com/materials-commons/mcstore/cmd/pkg/client"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/testdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	"github.com/willf/bitset"
	"net/http"
)

var _ = fmt.Println

var _ = Describe("UploadResource", func() {
	Describe("findStartingBlock method tests", func() {
		var (
			blocks *bitset.BitSet
		)

		BeforeEach(func() {
			blocks = bitset.New(10)
		})

		It("Should return 1 if no blocks have been set", func() {
			block := findStartingBlock(blocks)
			Expect(block).To(BeNumerically("==", 1))
		})

		It("Should return 2 if the first block has been uploaded", func() {
			// BitSet starts a zero. Flowjs starts at 1. So we have to adjust.
			blocks.Set(0)
			block := findStartingBlock(blocks)
			Expect(block).To(BeNumerically("==", 2))
		})

		It("Should return 1 if only the last block as been uploaded", func() {
			blocks.Set(9)
			block := findStartingBlock(blocks)
			Expect(block).To(BeNumerically("==", 1))
		})

		It("Should return 2 if only 2 has not been set (all others set)", func() {
			complement := blocks.Complement()
			complement.Clear(1) // second block
			block := findStartingBlock(complement)
			Expect(block).To(BeNumerically("==", 2))
		})
	})

	Describe("Upload REST API method tests", func() {
		var (
			client        *gorequest.SuperAgent
			server        *httptest.Server
			container     *restful.Container
			rr            *httptest.ResponseRecorder
			uploadRequest CreateUploadRequest
			uploads       dai.Uploads
		)

		BeforeEach(func() {
			client = c.NewGoRequest()
			container = NewServicesContainer(testdb.Sessions)
			server = httptest.NewServer(container)
			rr = httptest.NewRecorder()
			config.Set("mcurl", server.URL)
			uploadRequest = CreateUploadRequest{
				ProjectID:   "test",
				DirectoryID: "test",
				FileName:    "testreq.txt",
				FileSize:    4,
				ChunkSize:   2,
				FileMTime:   time.Now().Format(time.RFC1123),
				Checksum:    "abc123456",
			}
			uploads = dai.NewRUploads(testdb.RSessionMust())
		})

		var (
			createUploadRequest = func(req CreateUploadRequest) (*CreateUploadResponse, error) {
				r, body, errs := client.Post(Url("/upload")).Send(req).End()
				if err := ToError(r, errs); err != nil {
					return nil, err
				}

				var uploadResponse CreateUploadResponse
				if err := ToJSON(body, &uploadResponse); err != nil {
					return nil, err
				}
				return &uploadResponse, nil
			}
		)

		AfterEach(func() {
			server.Close()
		})

		Describe("create upload tests", func() {
			Context("No existing uploads that match request", func() {
				It("Should return an error when the user doesn't have permission", func() {
					// Set apikey for user who doesn't have permission
					config.Set("apikey", "test2")
					r, _, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).NotTo(BeNil())
					Expect(r.StatusCode).To(BeNumerically("==", http.StatusUnauthorized))
				})

				It("Should return an error when the project doesn't exist", func() {
					config.Set("apikey", "test")
					uploadRequest.ProjectID = "does-not-exist"
					r, _, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).NotTo(BeNil())
					Expect(r.StatusCode).To(BeNumerically("==", http.StatusNotFound))
				})

				It("Should return an error when the directory doesn't exist", func() {
					config.Set("apikey", "test")
					uploadRequest.DirectoryID = "does-not-exist"
					r, _, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).NotTo(BeNil())
					Expect(r.StatusCode).To(BeNumerically("==", http.StatusNotFound))
				})

				It("Should return an error when the apikey doesn't exist", func() {
					config.Set("apikey", "does-not-exist")
					r, _, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).NotTo(BeNil())
					Expect(r.StatusCode).To(BeNumerically("==", http.StatusUnauthorized))
				})

				It("Should create a new request for a valid submit", func() {
					config.Set("apikey", "test")
					r, body, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).To(BeNil())
					Expect(r.StatusCode).To(BeNumerically("==", http.StatusOK))
					var uploadResponse CreateUploadResponse
					err = ToJSON(body, &uploadResponse)
					Expect(err).To(BeNil())
					Expect(uploadResponse.StartingBlock).To(BeNumerically("==", 1))

					uploadEntry, err := uploads.ByID(uploadResponse.RequestID)
					Expect(err).To(BeNil())
					Expect(uploadEntry.ID).To(Equal(uploadResponse.RequestID))
					err = uploads.Delete(uploadEntry.ID)
					Expect(err).To(BeNil())
				})
			})

			Context("Existing uploads that could match", func() {
				var (
					idsToDelete []string
				)

				BeforeEach(func() {
					idsToDelete = []string{}
				})

				AfterEach(func() {
					for _, id := range idsToDelete {
						uploads.Delete(id)
					}
				})

				var addID = func(id string) {
					idsToDelete = append(idsToDelete, id)
				}

				It("Should find an existing upload rather than create a new one", func() {
					config.Set("apikey", "test")
					r, body, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).To(BeNil())
					var firstUploadResponse CreateUploadResponse
					err = ToJSON(body, &firstUploadResponse)
					Expect(err).To(BeNil())
					Expect(firstUploadResponse.StartingBlock).To(BeNumerically("==", 1))

					// Resend request - we should get the exact same request id back
					r, body, errs = client.Post(Url("/upload")).Send(uploadRequest).End()
					err = ToError(r, errs)
					Expect(err).To(BeNil())
					var secondUploadResponse CreateUploadResponse
					err = ToJSON(body, &secondUploadResponse)
					Expect(err).To(BeNil())
					Expect(secondUploadResponse.StartingBlock).To(BeNumerically("==", firstUploadResponse.StartingBlock))
					Expect(secondUploadResponse.RequestID).To(Equal(firstUploadResponse.RequestID))
					addID(firstUploadResponse.RequestID)
				})

				It("Should create a new upload when the request has a different checksum", func() {
					// Create two upload requests that are identical except for their checksums. This
					// should result in two different requests.

					config.Set("apikey", "test")
					r, body, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).To(BeNil())
					var firstUploadResponse CreateUploadResponse
					err = ToJSON(body, &firstUploadResponse)
					Expect(err).To(BeNil())
					Expect(firstUploadResponse.StartingBlock).To(BeNumerically("==", 1))
					addID(firstUploadResponse.RequestID)

					// Send second request with a different checksum
					uploadRequest.Checksum = "def456"
					r, body, errs = client.Post(Url("/upload")).Send(uploadRequest).End()
					err = ToError(r, errs)
					Expect(err).To(BeNil())
					var secondUploadResponse CreateUploadResponse
					err = ToJSON(body, &secondUploadResponse)
					Expect(err).To(BeNil())
					Expect(secondUploadResponse.StartingBlock).To(BeNumerically("==", 1))
					Expect(secondUploadResponse.RequestID).NotTo(Equal(firstUploadResponse.RequestID))
					addID(secondUploadResponse.RequestID)
				})
			})

			Context("Restarting upload requests", func() {
				var (
					idsToDelete []string
				)

				BeforeEach(func() {
					idsToDelete = []string{}
				})

				AfterEach(func() {
					for _, id := range idsToDelete {
						uploads.Delete(id)
					}
				})

				var addID = func(id string) {
					idsToDelete = append(idsToDelete, id)
				}

				It("Should ask for second block after sending first block and then requesting upload again", func() {
					config.Set("apikey", "test")
					r, body, errs := client.Post(Url("/upload")).Send(uploadRequest).End()
					err := ToError(r, errs)
					Expect(err).To(BeNil())
					Expect(r.StatusCode).To(BeNumerically("==", http.StatusOK))
					var uploadResponse CreateUploadResponse
					err = ToJSON(body, &uploadResponse)
					Expect(err).To(BeNil())
					Expect(uploadResponse.StartingBlock).To(BeNumerically("==", 1))
					addID(uploadResponse.RequestID)

					// Second first block
					ezclient := ezhttp.NewClient()
					params := make(map[string]string)
					params["flowChunkNumber"] = "1"
					params["flowTotalChunks"] = "2"
					params["flowChunkSize"] = "2"
					params["flowTotalSize"] = "4"
					params["flowIdentifier"] = uploadResponse.RequestID
					params["flowFileName"] = "testreq.txt"
					params["flowRelativePath"] = "test/testreq.txt"
					params["projectID"] = "test"
					params["directoryID"] = "test"
					params["fileID"] = ""
					sc, err, body := ezclient.PostFileBytes(Url("/upload/chunk"), "/tmp/test.txt", "chunkData",
						[]byte("ab"), params)
					Expect(err).To(BeNil())
					Expect(sc).To(BeNumerically("==", http.StatusOK))
					var chunkResp UploadChunkResponse
					err = ToJSON(body, &chunkResp)
					Expect(err).To(BeNil())
					Expect(chunkResp.Done).To(BeFalse())

					// Now we will request this upload a second time.
					r, body, errs = client.Post(Url("/upload")).Send(uploadRequest).End()
					err = ToError(r, errs)
					Expect(err).To(BeNil())
					Expect(r.StatusCode).To(BeNumerically("==", http.StatusOK))
					var uploadResponse2 CreateUploadResponse
					err = ToJSON(body, &uploadResponse2)
					Expect(err).To(BeNil())
					Expect(uploadResponse2.StartingBlock).To(BeNumerically("==", 2))
					Expect(uploadResponse2.RequestID).To(Equal(uploadResponse.RequestID))
				})

				It("Should return an error when sending a bad id", func() {
					ezclient := ezhttp.NewClient()
					params := make(map[string]string)
					params["flowChunkNumber"] = "1"
					params["flowTotalChunks"] = "2"
					params["flowChunkSize"] = "2"
					params["flowTotalSize"] = "4"
					params["flowIdentifier"] = "i-dont-exist"
					params["flowFileName"] = "testreq.txt"
					params["flowRelativePath"] = "test/testreq.txt"
					params["projectID"] = "test"
					params["directoryID"] = "test"
					params["fileID"] = ""
					_, err, _ := ezclient.PostFileBytes(Url("/upload/chunk"), "/tmp/test.txt", "chunkData",
						[]byte("ab"), params)
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Describe("get uploads tests", func() {
			It("Should return an error on a bad apikey", func() {
				config.Set("apikey", "test")
				resp, err := createUploadRequest(uploadRequest)
				Expect(err).To(BeNil())

				config.Set("apikey", "bad-key")
				r, _, errs := client.Get(Url("/upload/test")).End()
				err = ToError(r, errs)
				Expect(err).ToNot(BeNil())
				Expect(r.StatusCode).To(BeNumerically("==", http.StatusUnauthorized))

				err = uploads.Delete(resp.RequestID)
				Expect(err).To(BeNil())
			})

			It("Should return an error on a bad project", func() {
				config.Set("apikey", "test")
				r, _, errs := client.Get(Url("/upload/bad-project-id")).End()
				err := ToError(r, errs)
				Expect(err).ToNot(BeNil())
				Expect(r.StatusCode).To(BeNumerically("==", http.StatusBadRequest))
			})

			It("Should get existing upload requests for a project", func() {
				config.Set("apikey", "test")
				resp, err := createUploadRequest(uploadRequest)
				Expect(err).To(BeNil())
				r, body, errs := client.Get(Url("/upload/test")).End()
				err = ToError(r, errs)
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(BeNumerically("==", http.StatusOK))
				var entries []UploadEntry
				err = ToJSON(body, &entries)
				Expect(err).To(BeNil())
				Expect(entries).To(HaveLen(1))
				entry := entries[0]
				Expect(entry.RequestID).To(Equal(resp.RequestID))

				err = uploads.Delete(resp.RequestID)
				Expect(err).To(BeNil())
			})
		})
	})
})
