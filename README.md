# DOWNLOAD
* go run . http...

# DOWNLOAD with different name
* go run . -O=<new_name.type> http...

# DOWNLOAD in a standard file
* go run . -P=~/Downloads -O=<new_name.type> http...

# DOWNLOAD with limit speed
* go run . --rate-limit=200k http...

# DOWNLOAD a lot files at the same time from a file
* go run . -i=links.txt

# DOWNLOAD all the site
* go run . --mirror http...