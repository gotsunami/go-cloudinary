go-cloudinary
=============

A Go client library and CLI tool to upload static assets to the `Cloudinary`_ service.

.. _Cloudinary: http://www.cloudinary.com

Installation
------------

Install the CLI tool and the library with::

    go get github.com/gotsunami/go-cloudinary/cloudinary

Usage
-----

Usage::

    cloudinary [options] action settings.conf
    
where action is one of ``ls``, ``rm``, ``up`` or ``url``.

Create a config file ``settings.conf`` with a ``[cloudinary]`` section::

    [cloudinary]
    uri=cloudinary://api_key:api_secret@cloud_name

Type ``cloudinary`` in the terminal to get some help.

Uploading files
~~~~~~~~~~~~~~~

Use the ``up`` action to upload an image (or a directory of images) to Cloudinary with ``-i`` option::

    $ cloudinary -i /path/to/img.png up settings.conf
    $ cloudinary -i /path/to/images/ up settings.conf
    
In order to upload raw files, use the ``-r`` option. For example, a CSS file can be upload with::

    $ cloudinary -r /path/to/default.css up settings.conf

Raw files can be of any type (css, js, pdf etc.), even images if you don't
care about not using Cloudinary's image processing features.

List Remote Resources
~~~~~~~~~~~~~~~~~~~~~

Using the ``ls`` action will list all uploaded images and raw files::

    $ cloudinary ls settings.conf

Delete Remote Resources
~~~~~~~~~~~~~~~~~~~~~~~

Use the ``rm`` action to delete resources and give the ``-i`` or ``-r`` ``public_id`` for the resource::

    $ cloudinary -i img/home rm settings.conf
    $ cloudinary -r media/js/jquery-min.js rm settings.conf

Delete all remote resources(!) with::

    $ cloudinary -a rm settings.conf
    
In any case, you can always use the ``-s`` flag to simulate an action and see what result to expect.
i
