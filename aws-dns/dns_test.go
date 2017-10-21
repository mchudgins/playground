package awsdns

import (
	"fmt"
	"net"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func TestLookup(t *testing.T) {
	ip, err := net.LookupIP("www.dstsystems.com")
	if err != nil {
		t.Fatal(err)
		t.Fail()
	}
	if len(ip) != 1 {
		t.Fatal("expected 1 result from net.LookupIP('www.google.com') and got %d\n", len(ip))
		t.Fail()
	}
	if ip[0].String() != "162.209.114.126" {
		t.Fatal("expected net.LookupIP('www.google.com' to resolve to 162.209.114.126 and got %s\n", ip[0].String())
		t.Fail()
	}

	for _, addr := range ip {
		fmt.Printf("addr: %s\n", addr.String())
	}
}

func TestZoneLookup(t *testing.T) {
	sess := session.New()
	r53 := route53.New(sess)

	req := &route53.ListHostedZonesInput{}
	out, err := r53.ListHostedZones(req)
	if err != nil {
		t.Fatal(err)
	}

	for _, zone := range out.HostedZones {
		fmt.Printf("%-20s: %s\n", *zone.Name, *zone.Id)
	}
}

func TestZoneUpsert(t *testing.T) {
	sess := session.New()
	r53 := route53.New(sess)

	changeSet := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String("/hostedzone/Z3MB211DOB2GYS"),
		ChangeBatch: &route53.ChangeBatch{
			Comment: aws.String("updating"),
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String("mch-dev.dstcorp.io"),
						Type: aws.String("A"),
						TTL:  aws.Int64(300),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String("18.221.53.50"),
							},
						},
					},
				},
			},
		},
	}
	rc, err := r53.ChangeResourceRecordSets(changeSet)
	if err != nil {
		t.Fatal(err)
		t.Fail()
	}
	fmt.Printf("%s\n", rc)
}
