![Website](https://img.shields.io/website?label=vmweventbroker.io&url=https%3A%2F%2Fvmweventbroker.io%2F)

# Readme for the Website

## Structure

The website is hosted using [Github Pages](https://help.github.com/en/github/working-with-github-pages/about-github-pages) and built using [Jekyll](https://jekyllrb.com/). The files that make up the website are contained within the `docs` folder (as Github Pages requires) within the master branch. You'll find more details about how they are organized and their purpose below.

```
.
├── site                      > Contains MD files that need to go under the base website
│   └── **.md
├── kb                        > Contains MD files for the documentation section of the website
│   ├── img
│   │   └── **.png            > images required for documentation
│   └── **.md                 > All the MDs that make up the documentation
├── assets                    > Contains JS, CSS, IMGs for the site
│   ├── js
│   ├── img
│   └── css
├── index.html                > Website Landing page
├── README.md               *** You are here
├── _config.yml               > Site wide configuration and variables
└── Gemfile                   > Plugins required for the website to be built by Jekyll
```

In order for Jekyll to process the MD files and render them as html, you'll need to add the below to the beginning of the each MD file.

```yaml
---
layout: resources             # choose between default, docs, page or resources
title: Additional Resources   # provide the title for the web page
description: Update this      # this shows up in the Website description
permalink: /resources         # this is the short link for the page, if empty the relative path of the md file is used
#other yaml data that can be referenced within the page
---
```

### Other Key Files and Folders
- **_data/default.yml:** YAML content that drives the side-nav bar for the documentation
- **_data/resources.yml:** YAML content for the videos, links and external references contained in the resources page
- **_data/team.yml:** YAML data of the core team for the landing page
- **_functions:** folder that contains all the featured functions showcased on the landing page
- **_usecases:** folder that contains all the use cases showcased on the landing page
- **_includes** all the reusable html components referenced with the layouts
- **_layouts:** all the various layouts available to be used within the site
  - **docs** - use this for layout for the docs
  - **page** - use this for the pages that needs to go on the base site
  - **resources** - specifically designed for the resources page


## Run the website locally
To validate changes to any file/folder to the website, please verify them locally before you push to the repo.

### Pre-Reqs
* Install [Docker Client for your operating system](https://docs.docker.com/get-docker/)

### Build and View Documentation

Step 1 - Change into the `docs` directory

Step 2 - Run the following command to start the [Jekyll Docker container image](https://github.com/envygeeks/jekyll-docker/) and begin serving the documentation:

Linux/Mac:
```bash
docker run --rm \
  --volume="$PWD:/srv/jekyll" \
  --publish 4000:4000 \
  jekyll/jekyll:3.8 \
  jekyll serve
```

Windows:
```powershell
docker run --rm `
  --volume="${PWD}:/srv/jekyll" `
  --publish 4000:4000 `
  jekyll/jekyll:3.8 `
  jekyll serve
```
> Note: You may see a warning saying `Auto-regeneration may not work on some
> Windows versions.` Symptoms of this warning are the inability to automatically
> see any local website changes in your browser. This makes website changes
> difficult as you must restart Jekyll after every local change to view the
> results. To work around this issue, try the `--force-polling` switch to the
> `jekyll serve` command.
```powershell
docker run --rm `
  --volume="${PWD}:/srv/jekyll" `
  --publish 4000:4000 `
  jekyll/jekyll `
  jekyll serve --force-polling
```

Step 3 - Once the server is ready, you can open a browser to `http://localhost:4000` to review the documentation locally. If you need to change the default port (`4000`), modify the `--publish` arguments from step 2.

```bash
<snip>

Configuration file: /srv/jekyll/_config.yml
   GitHub Metadata: No GitHub API authentication could be found. Some fields may be missing or have incorrect data.
fatal: not a git repository (or any parent up to mount point /srv)
Stopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).
            Source: /srv/jekyll
       Destination: /srv/jekyll/_site
 Incremental build: disabled. Enable with --incremental
      Generating...
YAML Exception reading /srv/jekyll/site/examples-knative.md: (<unknown>): did not find expected key while parsing a block mapping at line 2 column 1
       Jekyll Feed: Generating feed for posts
                    done in 7.926 seconds.
 Auto-regeneration: enabled for '/srv/jekyll'
    Server address: http://0.0.0.0:4000
  Server running... press ctrl-c to stop.
```

**Note:** To stop serving the documentation using the Jekyll container, press Ctrl+C