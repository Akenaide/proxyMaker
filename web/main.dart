// Copyright (c) 2016, <your name>. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

import 'dart:html';
import 'dart:async';
import 'dart:convert';

final int powerHeight = 85;

removeImage(event) {
  event.target.remove();
}

allowDrop(Event event) {
  event.preventDefault();
}

drag(MouseEvent event) {
  event.dataTransfer.setData("text", event.target.id);
}

drop(MouseEvent event) {
  event.preventDefault();
  String data = event.dataTransfer.getData("text");
  print("hey");
  print(data);
  CanvasElement canvas = event.target;
  CanvasRenderingContext2D context = canvas.getContext("2d");
  CanvasElement sourceCanvas = querySelector("#"+data);
  double finalHeight = sourceCanvas.height * (canvas.width / sourceCanvas.width);
  context.drawImageScaled(sourceCanvas, 0, (canvas.height - finalHeight - powerHeight), canvas.width, finalHeight);
}

addImage(event) {

  var image = new ImageElement()
    ..src = event.target.src;

  CanvasElement canvas = new CanvasElement()
   ..classes.add("image")
   ..onDragOver.listen((e) => allowDrop(e))
   ..onDrop.listen((e) => drop(e));

   image.onLoad.listen((e) {
    CanvasRenderingContext2D context = canvas.getContext("2d");
    canvas
      ..width = image.width
      ..height = image.height;

    context.drawImage(image, 0, 0);
   });


  querySelector('#output').append(canvas);
  querySelectorAll('#output canvas.image').onClick.listen((e) => removeImage(e));
}

getCardImages() {
  InputElement url = querySelector("#url");
  var output = querySelector('#images-box');
  HttpRequest.postFormData("/cardimages", {"url": url.value}).then((HttpRequest response) {
    List parsedList = JSON.decode(response.response);
    for (var url in parsedList) {
      var image = new ImageElement();
      image.src = url;
      output.append(image);
    };
    querySelectorAll("#images-box img").onClick.listen((e) => addImage(e));
  });
}

Future loadImages(List<ImageElement> imgs) {
  List<Future> imageFutures = [];
  for (var i = 0; i < imgs.length; i++) {
    var img = imgs[i];
    imageFutures.add(img.onLoad.first);
  }

  return Future.wait(imageFutures);
}

initCanvas(List<ImageElement> imgs, CanvasElement canvas) {
  CanvasRenderingContext2D context = canvas.getContext('2d');
  int totalHeight;
  loadImages(imgs).then((images) {
    totalHeight = imgs[0].height * imgs.length;
    canvas
      ..width = imgs[0].width
      ..height = totalHeight;

    for (var i = 0; i < imgs.length; i++) {
      context.drawImage(imgs[i], 0, imgs[i].height * i);
    }
    querySelector("body").append(canvas);
  });
}

bindCanvas(CanvasElement canvas){
  List points = [];
  CanvasElement outCanvas = new CanvasElement();
  CanvasRenderingContext2D outContext = outCanvas.getContext('2d');
  canvas.onMouseDown.listen((e) {
    points.add(e.offset);
    if (points.length == 2) {
      outContext.clearRect(0, 0, outCanvas.width, outCanvas.height);
      Rectangle rect = new Rectangle.fromPoints(points[0], points[1]);
      outCanvas
        ..attributes.addAll({"draggable": true})
        ..id = "yay"
        ..classes.add("no-print")
        // ..id = new DateTime.now().millisecondsSinceEpoch.toString()
        ..onDragStart.listen((e) => drag(e))
        ..width = rect.width
        ..height = rect.height;
      outContext.drawImageToRect(canvas, new Rectangle(0, 0, rect.width, rect.height),
        sourceRect: rect);
      points.clear();
      querySelector("#output").append(outCanvas);
    }
  });
}

getPdfFile() {
  InputElement input = querySelector("#input-file");
  for (var file in input.files) {
    var reader = new FileReader();
    reader.onLoad.listen((e) {
      HttpRequest.postFormData("/translationimages", {"file": reader.result, "filename": file.name}).then((HttpRequest response) {
        List parsedList = JSON.decode(response.response);
        List images = [];
        CanvasElement canvas = new CanvasElement()
          ..classes.add("no-print");
        for (var url in parsedList) {
          ImageElement image = new ImageElement()
            ..src = url;
          images.add(image);
        };
        initCanvas(images, canvas);

        bindCanvas(canvas);
      });
    });
    reader.readAsDataUrl(file);
  }
}

void main() {
  querySelector("#input-file").onChange.listen((e) => getPdfFile());
  querySelector("#send-url").onClick.listen((e) => getCardImages());
  // querySelector("#yay").onDragStart.listen((e) => drag(e));

  // CanvasElement canvas = querySelector('#myCanvas');
  // ImageElement img = new ImageElement(src: "./love.png");
  // querySelector("body").append(img);
  // img.onLoad.listen((e) => initCanvas(img, canvas));

//   querySelector("#enterUrl").onclick.listen((event) {});
}
