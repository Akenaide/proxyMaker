// Copyright (c) 2016, <your name>. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

import 'dart:html';
import 'dart:convert';

removeImage(event) {
  event.target.remove();
}

allowDrop(Event event) {
  event.preventDefault();
}

drag(MouseEvent event) {
  print("YAy ====");
  event.dataTransfer.setData("text", event.target.id);
}

drop(MouseEvent event) {
  event.preventDefault();
  var data = event.dataTransfer.getData("text");
  print(data);
  print(querySelector("#"+data));
  event.target.append(querySelector("#"+data));
}

addImage(event) {
  var image = new ImageElement()
    ..src = event.target.src
    ..classes.add("image")
    ..onDragOver.listen((e) => allowDrop(e))
    ..onDrop.listen((e) => drop(e));

  querySelector('#output').append(image);
  querySelectorAll('#output img').onClick.listen((e) => removeImage(e));
}

getCardImages() {
  InputElement url = querySelector("#url").value;
  var output = querySelector('#images-box');
  HttpRequest.postFormData("/cardimages", {"url": url}).then((HttpRequest response) {
    List parsedList = JSON.decode(response.response);
    for (var url in parsedList) {
      var image = new ImageElement();
      image.src = url;
      output.append(image);
    };
    querySelectorAll("#images-box img").onClick.listen((e) => addImage(e));
  });
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
  querySelector("#input-dir").onChange.listen((e) => getPdfFile());
  querySelector("#send-url").onClick.listen((e) => getCardImages());
  querySelector("#yay").onDragStart.listen((e) => drag(e));

//   querySelector("#enterUrl").onclick.listen((event) {});
}
