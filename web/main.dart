// Copyright (c) 2016, <your name>. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

import 'dart:html';

getCardImages() {
  InputElement url = querySelector("#url").value;
  HttpRequest.postFormData("/cardimages", {"url": url}).then((HttpRequest response) {
    print(response);
  });

  // querySelector('#output').append(images);
}

getPdfFile() {
  print(querySelector("#input-dir").files);
  for (var file in querySelector("#input-dir").files) {
    var reader = new FileReader();
    reader.onLoad.listen((e) {
      var thumbnail = new Element.tag("embed");
      thumbnail.src = reader.result;
      querySelector('#output').append(thumbnail);
    });
    reader.readAsDataUrl(file);
  }
}

void main() {
  querySelector('#output').text = 'Your Dart app is running.';
  querySelector("#input-dir").onChange.listen((e) => getPdfFile());
  querySelector("#send-url").onClick.listen((e) => getCardImages());
//   querySelector("#enterUrl").onclick.listen((event) {});
}
