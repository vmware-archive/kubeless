/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"strconv"
)

var (
	// ErrorCrtFileMandatory when private key is provided but certificate not.
	ErrorCrtFileMandatory = errors.New("client certificate is mandatory when private key file is provided")
	// ErrorKeyFileMandatory when certificate is provided but private not.
	ErrorKeyFileMandatory = errors.New("client private key is mandatory when certificate file is provided")
	// ErrorUsernameOrPasswordMandatory when username or password is not provided
	ErrorUsernameOrPasswordMandatory = errors.New("username and password is mandatory")
)

// GetTLSConfiguration build TLS configuration for kafka, return tlsConfig associated with kafka connection.
func GetTLSConfiguration(caFile string, certFile string, keyFile string, insecure string) (*tls.Config, error) {

	if certFile != "" && keyFile == "" {
		return nil, ErrorKeyFileMandatory
	}
	if keyFile != "" && certFile == "" {
		return nil, ErrorCrtFileMandatory
	}

	isInsecure, _ := strconv.ParseBool(insecure)
	t := &tls.Config{}

	if caFile != "" {
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		t.RootCAs = caCertPool
	}

	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		t.Certificates = []tls.Certificate{cert}
	}

	t.InsecureSkipVerify = isInsecure
	return t, nil
}

// GetSASLConfiguration build SASL configuration for kafka.
func GetSASLConfiguration(username string, password string) (string, string, error) {
	if username == "" || password == "" {
		return "", "", ErrorUsernameOrPasswordMandatory
	}
	return username, password, nil
}
