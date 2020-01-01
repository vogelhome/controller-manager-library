/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package ctxutil

import (
	"context"
	"reflect"
)

type ValueType interface {
	Name() string
	WithValue(ctx context.Context, value interface{}) context.Context
	Get(ctx context.Context) interface{}
}

type valueType struct {
	name string
	key  reflect.Type
}

func NewValueType(name string, proto interface{}) ValueType {
	t := reflect.TypeOf(proto)
	return &valueType{
		name: name,
		key:  t,
	}
}

func (this *valueType) Name() string {
	return this.name
}

func (this *valueType) WithValue(ctx context.Context, value interface{}) context.Context {
	return context.WithValue(ctx, this.key, value)
}

func (this *valueType) Get(ctx context.Context) interface{} {
	return ctx.Value(this.key)
}
