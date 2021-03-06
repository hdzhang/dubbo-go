/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client_impl

import (
	"context"
	"net"
	"net/http"
	"path"
	"time"
)

import (
	"github.com/go-resty/resty/v2"
	perrors "github.com/pkg/errors"
)

import (
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/extension"
	"github.com/apache/dubbo-go/protocol/rest/client"
)

func init() {
	extension.SetRestClient(constant.DEFAULT_REST_CLIENT, NewRestyClient)
}

type RestyClient struct {
	client *resty.Client
}

func NewRestyClient(restOption *client.RestOptions) client.RestClient {
	client := resty.New()
	client.SetTransport(
		&http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, restOption.ConnectTimeout)
				if err != nil {
					return nil, err
				}
				err = c.SetDeadline(time.Now().Add(restOption.RequestTimeout))
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		})
	return &RestyClient{
		client: client,
	}
}

func (rc *RestyClient) Do(restRequest *client.RestRequest, res interface{}) error {
	r, err := rc.client.R().
		SetHeader("Content-Type", restRequest.Consumes).
		SetHeader("Accept", restRequest.Produces).
		SetPathParams(restRequest.PathParams).
		SetQueryParams(restRequest.QueryParams).
		SetHeaders(restRequest.Headers).
		SetBody(restRequest.Body).
		SetResult(res).
		Execute(restRequest.Method, "http://"+path.Join(restRequest.Location, restRequest.Path))
	if err != nil {
		return perrors.WithStack(err)
	}
	if r.IsError() {
		return perrors.New(r.String())
	}
	return nil
}
