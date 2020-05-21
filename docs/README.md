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


## Build and Run the website locally
To ensure the changes to any file or folder that power the website is valid, please setup this step below that allows you to build the website, verify changes locally before you push to the repo.

### Dependencies for MacOS

Install the following for an easy to use dev environment:

```bash
brew install rbenv
rbenv install 2.6.3
gem install bundler
```

*Note: if you hit a permissions error for the `gem install bundler` follow advice from the [bundler docs](https://bundler.io/doc/troubleshooting.html#permission-denied-when-installing-bundler)*

### Dependencies for Linux
If you are running a build on Ubuntu you will need the following packages:
* ruby
* ruby-dev
* ruby-bundler
* build-essential
* zlib1g-dev
* nginx (or apache2)

### Dependencies for Windows
If you are on Windows, all hope is not lost. Follow the steps here to install the dependencies - https://jekyllrb.com/docs/installation/windows/

### Local Development
* Install Jekyll and plug-ins in one fell swoop. `gem install github-pages`
This mirrors the plug-ins used by GitHub Pages on your local machine including Jekyll, Sass, etc.
* Clone down your own fork, or clone the main repo and add your own remote.

```bash
git clone git@github.com:vmware-samples/vcenter-event-broker-appliance.git
cd vcenter-event-broker-appliance/docs
bundle install
```

* Serve the site and watch for markup/sass changes `jekyll serve --livereload --incremental`. You may need to run `bundle exec jekyll serve --livereload --incremental`.
* View your website at http://127.0.0.1:4000/
* Commit any changes and push everything to your fork.
* Once you're ready, submit a PR of your changes.

## Troubleshooting
* If you don't see your updates reflected on the website when running locally, try the following steps

```zsh
bundle exec jekyll clean
bundle exec jekyll serve --incremental --livereload
```
