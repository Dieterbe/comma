Comma: a super simple comment server for static websites, in go

This means your website is no longer purely static, but it's a simple, fairly pragmatic solution to dynamically save and load comments (using javascript)

# API

* GET /foo/bar/whatever/my-post-slug
  provides all comments for post `my-post-slug` as JSON.

* POST
  reads out the following form fields, creates a comment and returns the comment as JSON.
  - post
  - message
  - name
  - email
  - url
  - company (honey-pot field to prevent spam, must be empty)

# Storage

Files are stored in xml files compatible with [pyblosxom's comment files](http://pyblosxom.github.io/)

email addresses and ip addresses of comments are never served up, though the md5 of email addresses is,
so you can use gravatar.

# Html/javascript based commenting feature on top of this server

Integrating this in your website takes less than 100 lines of javascript.
You can either:

* use jquery, like [on my old blog](https://github.com/Dieterbe/hugo-theme-blog/blob/master/layouts/partials/comments.html)
* use pure javascript, like [on my new blog](https://github.com/Dieterbe/dieterblog/blob/master/layouts/partials/comments.html)


See it in action on [dieter.plaetinck.be](http://dieter.plaetinck.be/)

## How to run

I use a systemd unit like this:
```
[Unit]
Description=comma backend
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/home/dieter/comma /home/dieter/<comments directory> :<port> [form value for "special" form fields]
Restart=always
RestartSec=1
User=dieter
Group=dieter

[Install]
WantedBy=graphical.target
```

the optional "special" argument is a basic spam prevention mechanism. if the value is provided, and the "special" form value doesn't match this value,
the comment is rejected. 
