

This is WP for the reporting for crowbar-server.
Part of my HACKWEEK project.

Also a chance to practice go and write a modular webserver in go.
I did not write this from scratch rather stole it from github repo and then stripped down to bare bones.


Usage:

1.
Add this to the client.rb file of crowbar, this is the config file in the /etc dir
```
enable_reporting true
```



2
Patch the `chef-server-api-10.32.2/config/router.rb` file
```
  # Reports
  match('/reports/nodes/:id/runs', :id => /[^\/]+/, :method => 'post').
    to(:controller => "reports", :action => "newrun")
  match('/reports/nodes/:id/runs/:runid', :id => /[^\/]+/, :runid => /[^\/]+/, :method => 'post').
    to(:controller => "reports", :action => "stoprun")
```


3.
Add a file `chef-server-api-10.32.2/app/controllers/reports.rb`
```
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
```


4.
Run a binary of the go project in this repo, it start a server on port 8123 
(have a config.json, it is needed to run the server)


./crowbar-stats --config path-to-config.json-file


# What, Where and Huhhh...


```
chef-client :
    begin-run: tell server about it and get ID fot this run for server
      end-run: tell server about the end and post run-times as gzipped-json to the server

Server :
    begin-run:
              - generate a ID
              - tie it to the chef-client of the node by storing it in DB
              - return it to the client
     end-run:
              - get the ID from post request and validate it in DB
              - stored the gzipped-json in a folder (data-folder)
              - mark run completion in the DB
```

Name of DB and data-folder: 
    the go server has a config file where these names and paths can be defined, look config.json
