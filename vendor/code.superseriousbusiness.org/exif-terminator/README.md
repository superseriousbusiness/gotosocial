# exif-terminator

`exif-terminator` removes exif data from images (jpeg and png currently supported) in a streaming manner. All you need to do is provide a reader of the image in, and exif-terminator will provide a reader of the image out.

Hasta la vista, baby!

```text
                                                  .,lddxococ.                   
                                                 ..',lxO0Oo;'.                  
                                               .  .. .,coodO0klc:.              
                           .,.   ..','.  ..  .,..'.      .':llxKXk'             
                          .;c:cc;;,...   .''.,l:cc. .....:l:,,:oo:..            
                          .,:ll'.   .,;cox0OxOKKXX0kOOxlcld0X0d;,,,'.           
                          .:xkl. .':cdKNWWWWMMMMMMMMMMWWNXK0KWNd.               
                         .coxo,..:ollk0KKXNWMMMMMMMMMMWWXXXOoOM0;               
                         ,oc,.  .;cloxOKXXWWMMMMMMMMMMMWNXk;;OWO'               
                          .      ..;cdOKXNNWWMMMMMMMMMMMMWO,,ONO'               
          ......                ....;okOO000XWWMMMMMMMMMWXx;,ONNx.              
.;c;.     .:l'ckl.              ..';looooolldolloooodolcc:;'.;oo:.              
.oxl.      ;:..OO.              .. ..             .,'         .;.               
.oko.     .cc.'Ok.                                .:;     .:,..';.              
.cdc.   .;;lc.,Ox.              .   .',,'..','.  .dN0; .. .c:,,':.              
.:oc.   ,dxkl.,0x.              .     ..   .    .oNMMKc..   ...:l.              
.:o:.   cKXKl.,Ox.              ..             .lKWMMMXo,.  ...''.              
.:l;    c0KKo.,0x.               ...........';:lk0OKNNXKkl,..,;cxd'             
.::'    ;k00l.;0d.        ..     .,cloooddddxxddol;:ddloxdc,:odOWNc             
.;,.    ,ONKc.;0d.        'l,..   .:clllllllokKOl::cllclkKx'.lolxx'             
.,.     '0W0:.;0d.        .:l,.   .,:ccc:::oOXNXOkxdook0NWNx,,;c;.              
...     .kX0c.;0d.         .loc'  .,::;;;;lk0kddoooooddooO0o',ld;               
..      .oOkk:cKd.          ....  .;:,',;cxK0o::ldkOkkOkxod:';oKx.              
..       :dlOolKO,                '::'.';:oOK0xdddoollooxOx::ccOx.              
..       ';:o,.xKo.               .,;'...';lddolooodkkkdol:,::lc.               
..       ...:..oOl.                ........';:codxxOXKKKk;':;:kl                
..         .,..lOc.               ..     ....,codxkxxxxxo:,,;lKO.  .,;'..       
...         .. ck:                ';,'.       .;:cllloc,;;;colOK;  .;odxxoc;.   
...,....    .  :x;                .;:cc;'.     .,;::c:'..,kXk:xNc   .':oook00x:.
      .        cKx.    .'..        ':clllc,...'';:::cc:;.,kOo:xNx.    .'codddoox
      ..       ,xxl;',col:;.       .:cccccc;;;:lxkkOOkdc,,lolcxWO'       ;kNKc.'
     .,.       .c' ':dkO0O;     .. .;ccccccc:::cldxkxoll:;oolcdN0:..      .xWNk;
     .:'       .c',xXNKkOXo    .,. .,:cccccllc::lloooolc:;lo:;oXKc,::.     .kWWX
      ,'       .cONMWMWkco,    ',  .';::ccclolc:llolollcccodo;:KXl..cl,.    ;KWN
      '.       .xWWWWMKc;; ....;'   ',;::::coolclloooollc:,:o;;0Xx, .,:;... ,0Ko
      .        ,kKNWWXd,cdd0NXKk:,;;;'';::::coollllllllllc;;ccl0Nkc.   ..';loOx'
               'lxXWMXOOXNMMMMWWNNNWXkc;;;;;:cllccccccccc::lllkNWXd,.   .cxO0Ol'
               ,xKNWWXkkXWM0dxKNWWWMWNX0OOkl;;:c::cccc:,...:oONMMXOo;.  :kOkOkl;
               .;,;:;...,::.  .;lokXKKNMMMWNOc,;;;,::;'...lOKNWNKkol:,..cKdcO0do
                       .:;...  .. .,:okO0KNN0:.',,''''. ':xNMWKkxxOKXd,.cNk,:l:o
```

## Why?

Exif removal is a pain in the arse. Most other libraries seem to parse the whole image into memory, then remove the exif data, then encode the image again.

`exif-terminator` differs in that it removes exif data *while scanning through the image bytes*, and it doesn't do any reencoding of the image. Bytes of exif data are simply all set to 0, and the image data is piped back out again into the returned reader.

The only exception is orientation data: if an image contains orientation data, this and only this data will be preserved since it's *actually useful*.

## Example

You can run the following example with `go run ./example/main.go`:

```go
package main

import (
  "io"
  "os"

  terminator "code.superseriousbusiness.org/exif-terminator"
)

func main() {
  // open a file
  sloth, err := os.Open("./images/sloth.jpg")
  if err != nil {
    panic(err)
  }
  defer sloth.Close()

  // terminate!
  out, err := terminator.Terminate(sloth, "jpeg")
  if err != nil {
    panic(err)
  }

  // read the bytes from the reader
  b, err := io.ReadAll(out)
  if err != nil {
    panic(err)
  }

  // save the file somewhere
  if err := os.WriteFile("./images/sloth-clean.jpg", b, 0666); err != nil {
    panic(err)
  }
}
```

## Credits

### Libraries

`exif-terminator` borrows heavily from the [`dsoprea`](https://github.com/dsoprea) libraries credited below. In fact, it's basically a hack on top of those libraries. Thanks `dsoprea`!

- [superseriousbusiness/go-jpeg-image-structure](https://code.superseriousbusiness.org/go-jpeg-image-structure): jpeg structure parsing. [MIT License](https://spdx.org/licenses/MIT.html). Forked from [dsoprea/go-jpeg-image-structure](https://github.com/dsoprea/go-jpeg-image-structure): jpeg structure parsing. [MIT License](https://spdx.org/licenses/MIT.html).
- [superseriousbusiness/go-png-image-structure](https://code.superseriousbusiness.org/go-png-image-structure): png structure parsing. Forked from [dsoprea/go-png-image-structure](https://github.com/dsoprea/go-png-image-structure): png structure parsing. [MIT License](https://spdx.org/licenses/MIT.html).
- [dsoprea/go-exif](https://github.com/dsoprea/go-exif): exif header reconstruction. [MIT License](https://spdx.org/licenses/MIT.html).
- [stretchr/testify](https://github.com/stretchr/testify); test framework. [MIT License](https://spdx.org/licenses/MIT.html).

## License

![the gnu AGPL logo](https://www.gnu.org/graphics/agplv3-155x51.png)

`exif-terminator` is free software, licensed under the [GNU AGPL v3 LICENSE](LICENSE).

Copyright (C) 2022-2025 SuperSeriousBusiness.
