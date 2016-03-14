// Copyright (c) 2016, <your name>. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

import 'dart:html';

import 'package:http/http.dart' as http;
import 'package:html/parser.dart' show parse;

// getCardImages(String url) {
//   HttpRequest.requestCrossOrigin(url).then((HttpRequest response) {
//     var document = parse(response.body);
//     var images = document.querySelectorAll(".card_list_box img");
//   });

//   querySelector('#output').append(images);
// }

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
//   querySelector("#enterUrl").onclick.listen((event) {});
}
