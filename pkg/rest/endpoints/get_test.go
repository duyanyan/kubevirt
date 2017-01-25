package endpoints

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"

	"encoding/json"
	"github.com/emicklei/go-restful"
	"github.com/ghodss/yaml"
	"golang.org/x/net/context"
	"gopkg.in/ini.v1"
	"kubevirt.io/kubevirt/pkg/rest"
	"net/http/httptest"
	"net/url"
)

func newValidGetRequest() *http.Request {
	request, _ := http.NewRequest("GET", "/apis/kubevirt.io/v1alpha1/namespaces/default/vms/test", nil)
	return request
}

func testGetEndpoint(_ context.Context, request interface{}) (interface{}, error) {
	metadata := request.(*Metadata)
	Expect(metadata.Name).To(Equal("test"))
	return &payload{Name: "test", Email: "test@test.com"}, nil
}

var _ = Describe("Get", func() {
	var recorder *httptest.ResponseRecorder
	var request *http.Request
	var handler http.Handler
	ctx := context.Background()

	BeforeEach(func() {

		ws := new(restful.WebService)
		ws.Produces(restful.MIME_JSON)
		handler = http.Handler(restful.NewContainer().Add(ws))

		target := MakeGoRestfulWrapper(NewHandlerBuilder().Get().Endpoint(testGetEndpoint).
			Encoder(NewMimeTypeAwareEncoder(NewEncodeJsonResponse(http.StatusOK), map[string]kithttp.EncodeResponseFunc{
				rest.MIME_TEXT: NewEncodeINIResponse(http.StatusOK),
				rest.MIME_YAML: NewEncodeYamlResponse(http.StatusOK),
			})).Build(ctx))
		ws.Route(ws.GET("/apis/kubevirt.io/v1alpha1/namespaces/{namespace}/vms/{name}").Produces(rest.MIME_JSON, rest.MIME_TEXT, rest.MIME_YAML).To(target))

		request = newValidGetRequest()
		recorder = httptest.NewRecorder()
	})

	Describe("REST call", func() {
		Context("with invalid URL", func() {
			It("should return 404", func() {
				request.URL, _ = url.Parse("/api/rest/put/")
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
		Context("with valid request", func() {
			It("should return 200", func() {
				handler.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
			It("should return a json containing the right name and email", func() {
				handler.ServeHTTP(recorder, request)
				responseBody := payload{}
				json.NewDecoder(recorder.Body).Decode(&responseBody)
				Expect(recorder.Header().Get("Content-Type")).To(Equal(rest.MIME_JSON))
				Expect(responseBody).To(Equal(payload{Name: "test", Email: "test@test.com"}))
			})
		})
		Context("with Accept header rest.MIME_TEXT", func() {
			It("should return ini file", func() {
				request.Header.Add("Accept", rest.MIME_TEXT)
				handler.ServeHTTP(recorder, request)
				responseBody := payload{}
				f, err := ini.Load(recorder.Body.Bytes())
				Expect(err).To(BeNil())
				Expect(recorder.Header().Get("Content-Type")).To(Equal(rest.MIME_TEXT))
				Expect(f.MapTo(&responseBody)).To(BeNil())
				Expect(responseBody).To(Equal(payload{Name: "test", Email: "test@test.com"}))
			})
		})
		Context("with Accept header applicatoin/yaml", func() {
			It("should return yaml file", func() {
				request.Header.Add("Accept", rest.MIME_YAML)
				handler.ServeHTTP(recorder, request)
				responseBody := payload{}
				Expect(recorder.Header().Get("Content-Type")).To(Equal(rest.MIME_YAML))
				Expect(yaml.Unmarshal(recorder.Body.Bytes(), &responseBody)).To(BeNil())
				Expect(responseBody).To(Equal(payload{Name: "test", Email: "test@test.com"}))
			})
		})
	})
})
