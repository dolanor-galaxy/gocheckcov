#!/bin/bash

license=$(cat << END
// Copyright © 2019 Cole Giovannoni Wippern
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

END
)
fail=0
for f in $(find . -type f | grep -e ".go$" | grep -v vendor); do
  if !(cat $f | grep "$license" -q); then
    echo $f 'does not have a license'
    fail=1
  fi
done

if [ $fail -eq 1 ]; then
  exit 1
fi
