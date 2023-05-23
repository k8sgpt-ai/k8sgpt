package util

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func (k *K8sApiReference) GetApiDoc(field string) (string, error) {
	client := resty.New()
	splitApiVersion := strings.Split(
		strings.ToLower(k.ApiVersion),
		"/",
	)

	resp, err := client.R().
		EnableTrace().
		Get(
			fmt.Sprintf(
				"https://kubernetes.io/docs/reference/generated/kubernetes-api/v%s.%s/#%sspec-%s-%s",
				k.ServerVersion.Major,
				k.ServerVersion.Minor,
				strings.ToLower(k.Kind),
				splitApiVersion[1],
				splitApiVersion[0],
			),
		)
	if err != nil {
		fmt.Printf("%w", err)
	}

	fieldDoc := ""
	isFound := false
	for _, line := range strings.Split(strings.TrimRight(string(resp.Body()), "\n"), "\n") {
		if strings.Contains(line, field) {
			tkn := html.NewTokenizer(strings.NewReader(line))
			isTd := false

			for {
				tt := tkn.Next()

				switch tt {
				case html.ErrorToken:
					break

				case html.StartTagToken:
					t := tkn.Token()
					isTd = t.Data == "td"

				case html.TextToken:
					t := tkn.Token()

					if isTd && t.Data != field {
						fieldDoc = t.Data
						isFound = true
					}

					isTd = false
				}

				if isFound {
					break
				}
			}
		}

		if isFound {
			break
		}
	}

	return fieldDoc, nil
}
