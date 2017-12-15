#
# Author:: Adam Jacob (<adam@opscode.com>)
# Author:: Christopher Brown (<cb@opscode.com>)
# Copyright:: Copyright (c) 2008 Opscode, Inc.
# License:: Apache License, Version 2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#


require "rest-client"

class Reports < Application

  provides :json

  before :authenticate_every

  def newrun
      r = RestClient.post("http://127.0.0.1:8123/reports/nodes/#{params[:id]}/runs/", {:action => :begin}.to_json)
      display Chef::JSONCompat.from_json(r.body) 
  end
  def stoprun
      RestClient.post("http://127.0.0.1:8123/reports/nodes/#{params[:id]}/runs/#{params[:runid]}", request.raw_post, {:content_type => "gzip"})
  end
end
