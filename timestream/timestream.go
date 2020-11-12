package timestream

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamquery"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"golang.org/x/net/http2"
	"log"
	"net"
	"net/http"
	"time"
)

type Timestream struct {
	Database string
	Table string
	session *session.Session
	writeSvc *timestreamwrite.TimestreamWrite
}

func New(database string, table string, region string, credentials *credentials.Credentials) Timestream {
	t := Timestream{Database: database, Table: table}

	tr := &http.Transport{
		ResponseHeaderTimeout: 20 * time.Second,
		// Using DefaultTransport values for other parameters: https://golang.org/pkg/net/http/#RoundTripper
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// So client makes HTTP/2 requests
	http2.ConfigureTransport(tr)

	var err error
	t.session, err = session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials,
		MaxRetries: aws.Int(10),
		HTTPClient: &http.Client{ Transport: tr },
	})
	if err != nil {
		log.Fatalln(err)
	}

	t.writeSvc = timestreamwrite.New(t.session)

	return t
}

func (t Timestream) Save(records []*timestreamwrite.Record) {
	writeRecordsInput := &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(t.Database),
		TableName:    aws.String(t.Table),
		Records: records,
	}

	_, err := t.writeSvc.WriteRecords(writeRecordsInput)

	if err != nil {
		log.Fatalln(err)
	}
}

func processScalarType(data *timestreamquery.Datum) string {
	return *data.ScalarValue
}

func (t Timestream) RunQuery(query string) string {
	querySvc := timestreamquery.New(t.session)

	queryInput := &timestreamquery.QueryInput{
		QueryString: aws.String(query),
	}

	out, err := querySvc.Query(queryInput)
	if err != nil {
		panic(err)
	}

	if len(out.Rows) == 0 {
		return ""
	}

	return processScalarType(out.Rows[0].Data[0])
}
