// Copyright (c) 2016, <your name>. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

import 'dart:html';
import 'dart:async';
import 'dart:convert';

final int powerHeight = 90;
DivElement spinner;

removeImage(event) {
  event.target.remove();
}

void toggleHideTranslation() {
  CanvasElement canvas = querySelector(".translation-canvas");
  canvas.classes.toggle("hide");
}

allowDrop(Event event) {
  event.preventDefault();
}

drag(MouseEvent event) {
  // event.dataTransfer.setData("text", event.target);
  print("Not implemented anymore");
}

drop(MouseEvent event) {
  event.preventDefault();
  String data = event.dataTransfer.getData("text");
  CanvasElement canvas = event.target;
  CanvasRenderingContext2D context = canvas.getContext("2d");
  // CanvasElement sourceCanvas = querySelector(data);
  CanvasElement sourceCanvas = querySelector("#" + data);
  double finalHeight =
      sourceCanvas.height * (canvas.width / sourceCanvas.width);
  context.drawImageScaled(sourceCanvas, 0, (canvas.height - powerHeight),
      canvas.width, finalHeight);
}

addImage(event) {
  var image = new ImageElement()..src = event.target.src;

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
  querySelectorAll('#output canvas.image')
      .onClick
      .listen((e) => removeImage(e));
}

getCardImages() {
  spinner.classes.toggle("hide");
  InputElement url = querySelector("#url");
  var control = querySelector('#right-panel');
  DivElement output = new DivElement()..id = "images-box";
  HttpRequest.postFormData("/cardimages", {"url": url.value}).then(
      (HttpRequest response) {
    List parsedList = JSON.decode(response.response);
    for (var url in parsedList) {
      var image = new ImageElement();
      image.src = url;
      output.append(image);
    }
    ;
    control.append(output);
    querySelectorAll("#images-box img").onClick.listen((e) => addImage(e));
    spinner.classes.toggle("hide");
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

checkReturnLine(String text) {}

getLines(String text) {
  int maxlength = 40;
  List<String> lines = [];
  String line = "";
  for (var word in text.split(" ")) {
    if (line.length <= maxlength) {
      if (word.contains("\n")) {
        var subline = word.split("\n");
        line = line + " " + subline[0];
        lines.add(line);
        line = subline[1];
      } else {
        line = line + " " + word;
      }
    } else {
      if (word.contains("\n")) {
        var subline = word.split("\n");
        line = line + " " + subline[0];
        lines.add(line);
        line = subline[1];
      } else {
        line = line + " " + word;
        lines.add(line);
        line = "";
      }
    }
  }
  if (line != "") {
    lines.add(line);
  }
  return lines;
}

bindCanvas(CanvasElement canvas) {
  List points = [];
  CanvasElement outCanvas = new CanvasElement();
  CanvasRenderingContext2D outContext = outCanvas.getContext('2d');
  canvas.onMouseDown.listen((e) {
    points.add(e.offset);
    if (points.length == 2) {
      outContext.clearRect(0, 0, outCanvas.width, outCanvas.height);
      Rectangle rect = new Rectangle.fromPoints(points[0], points[1]);
      outCanvas
        ..attributes.addAll({"draggable": "true"})
        ..id = "yay"
        ..classes.add("no-print")
        // ..id = new DateTime.now().millisecondsSinceEpoch.toString()
        ..onDragStart.listen((e) => drag(e))
        ..width = rect.width
        ..height = rect.height;
      outContext.drawImageToRect(
          canvas, new Rectangle(0, 0, rect.width, rect.height),
          sourceRect: rect);
      points.clear();
      querySelector("#output").append(outCanvas);
      querySelector(".translation-canvas").classes.toggle("hide");
    }
  });
}

getPdfFile() {
  InputElement input = querySelector("#input-file");
  for (var file in input.files) {
    var reader = new FileReader();
    reader.onLoad.listen((e) {
      HttpRequest.postFormData("/translationimages", {
        "file": reader.result,
        "filename": file.name
      }).then((HttpRequest response) {
        List parsedList = JSON.decode(response.response);
        List images = [];
        CanvasElement canvas = new CanvasElement()
          ..classes.add("translation-canvas")
          ..classes.add("hide")
          ..classes.add("no-print");
        for (var url in parsedList) {
          ImageElement image = new ImageElement()..src = url;
          images.add(image);
        }
        ;
        // initCanvas(images, canvas);

        bindCanvas(canvas);
      });
    });
    reader.readAsDataUrl(file);
  }
}

printTranslation() {
  spinner.classes.toggle("hide");
  InputElement deckUrl = querySelector("#url");

  HttpRequest.postFormData("/translationimages", {"url": deckUrl.value}).then(
      (HttpRequest response) {
    List parsedList = JSON.decode(response.response);
    for (var card in parsedList) {
      DivElement printDiv = new DivElement();
      Element cardId = new Element.tag("h3");
      ParagraphElement translation = new ParagraphElement();
      translation.appendHtml(card["Translation"].replaceAll("\n", "<br>"));
      cardId.appendText(card["ID"]);
      printDiv.append(cardId);
      printDiv.append(translation);

      querySelector('#output').append(printDiv);
    }
    spinner.classes.toggle("hide");
  });
}

estimatePrice() {
  spinner.classes.toggle("hide");
  InputElement deckUrl = querySelector("#url");
  TableElement table = new TableElement();
  table.classes.add("table");
  table.classes.add("table-bordered");
  TableSectionElement tbody = table.createTBody();
  TableSectionElement thead = table.createTHead();
  TableRowElement headRow = thead.addRow();
  headRow.addCell().text = "ID";
  headRow.addCell().text = "Image";
  headRow.addCell().text = "Price";
  headRow.addCell().text = "Amount";
  headRow.addCell().text = "Total";

  HttpRequest.postFormData("/estimateprice", {"url": deckUrl.value}).then(
      (HttpRequest response) {
    List parsedList = JSON.decode(response.response);
    parsedList.sort((a, b) => a["ID"].compareTo(b["ID"]));
    for (var card in parsedList) {
      ImageElement image = new ImageElement(src: card["URL"]);
      image.classes.add("estimate__image");
      TableRowElement row = tbody.addRow();
      row.addCell().text = card["ID"];
      row.addCell().children = [image];
      row.addCell().text = card["Price"].toString();
      row.addCell().text = card["Amount"].toString();
      row.addCell().text = card["Total"].toString();
    }
    querySelector('#output').append(table);
    spinner.classes.toggle("hide");
  });
}

void main() {
  querySelector("#input-file").onChange.listen((e) => getPdfFile());
  querySelector("#send-url").onClick.listen((e) => getCardImages());
  querySelector("#print-translation").onClick.listen((e) => printTranslation());
  querySelector("#estimate-price").onClick.listen((e) => estimatePrice());
  querySelector("#toggle-hide-translation")
      .onClick
      .listen((e) => toggleHideTranslation());
  spinner = querySelector("#spinner");
  // querySelector("#yay").onDragStart.listen((e) => drag(e));

  // CanvasElement canvas = querySelector('#myCanvas');
  // ImageElement img = new ImageElement(src: "./love.png");
  // querySelector("body").append(img);
  // img.onLoad.listen((e) => initCanvas(img, canvas));

//   querySelector("#enterUrl").onclick.listen((event) {});
}
