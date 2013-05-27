go-cloudinary
=============

A Go library and CLI tool to upload static assets to the `Cloudinary`_ service.

.. _Cloudinary: http://www.cloudinary.com

Installation
------------

Install the CLI tool and the library with::

    go get github.com/matm/go-cloudinary/cloudinary

Usage
-----

Create a config file ``settings.conf`` with a ``[cloudinary]`` section::

    [cloudinary]
    uri=cloudinary://api_key:api_secret@cloud_name

Type ``cloudinary`` in the terminal to get some help.

Uploading Images
~~~~~~~~~~~~~~~~

Use the ``-upi`` flag to upload an image (or a directory of images) to Cloudinary::

    $ cloudinary -upi /path/to/img.png settings.conf
    $ cloudinary -upi /path/to/images/ settings.conf

Uploading Raw Files
~~~~~~~~~~~~~~~~~~~

Use the ``-upr`` flag to upload files as raw files (no image processing at all) to Cloudinary::

    $ cloudinary -upr /path/to/img.png settings.conf
    $ cloudinary -upr /path/to/css/ settings.conf

Raw files can be of any type (css, js, pdf etc.), even images if you don't
care about not using Cloudinary's image processing features.

List Remote Images
~~~~~~~~~~~~~~~~~~

Use the ``-lsi`` flag for listing remote images::

    $ cloudinary -lsi settings.conf

List Remote Raw Files
~~~~~~~~~~~~~~~~~~~~~

Use the ``-lsr`` flag for listing remote raw files::

    $ cloudinary -lsr settings.conf

Delete Remote Images
~~~~~~~~~~~~~~~~~~~~

Use the ``-rmi`` flag to delete a remote image by ``public_id``::

    $ cloudinary -rmi img/home settings.conf

You may want to use the ``-rmalli`` flag to delete all remote images::

    $ cloudinary -rmalli settings.conf

Delete Remote Raw Files
~~~~~~~~~~~~~~~~~~~~~~~

Use the ``-rmr`` flag to delete a remote image by ``public_id``::

    $ cloudinary -rmr css/default.css settings.conf

You may want to use the ``-rmallr`` flag to delete all remote images::

    $ cloudinary -rmallr settings.conf
