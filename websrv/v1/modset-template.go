// Copyright 2015 lessOS.com, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessids/idclient"
	"github.com/lessos/lessids/idsapi"

	"../../api"
	"../../config"
	"../../modset"
)

func (c ModSet) FsTplListAction() {

	ls := api.ViewList{}

	defer c.RenderJson(&ls)

	if !idclient.SessionAccessAllowed(c.Session, "sys.admin", config.Config.InstanceID) {
		ls.Error = &types.ErrorMeta{idsapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	spec, err := modset.SpecFetch(c.Params.Get("modname"))
	if err != nil {
		ls.Error = &types.ErrorMeta{api.ErrCodeBadArgument, "ModName Not Found"}
		return
	}

	basepath := config.Config.Prefix + "/modules/" + spec.Meta.Name + "/views/"
	_ = filepath.Walk(basepath, func(path string, info os.FileInfo, err error) error {

		path = strings.TrimPrefix(path, basepath)

		if len(path) > 4 && path[len(path)-4:] == ".tpl" {
			ls.Items = append(ls.Items, api.View{
				Path: path,
			})
		}

		return nil
	})

	ls.Kind = "SpecTemplateList"
}