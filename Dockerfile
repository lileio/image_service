FROM marcbachmann/libvips

ADD build/image_server /bin
CMD ["image_server", "server"]
