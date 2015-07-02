Comma: a super simple comment server for static websites, in go

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

[Here's how I do it on my blog](https://github.com/Dieterbe/hugo-theme-blog/blob/master/layouts/partials/comments.html)

See it in action on [dieter.plaetinck.be](http://dieter.plaetinck.be/)
