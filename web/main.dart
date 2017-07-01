// Copyright (c) 2016, <your name>. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

import 'dart:html';
import 'dart:svg' as svg;
import 'dart:async';
import 'dart:convert';

final int powerHeight = 90;

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
  event.dataTransfer.setData("text", event.target.id);
}

drop(MouseEvent event) {
  event.preventDefault();
  String data = event.dataTransfer.getData("text");
  CanvasElement canvas = event.target;
  CanvasRenderingContext2D context = canvas.getContext("2d");
  // CanvasElement sourceCanvas = querySelector(data);
  CanvasElement sourceCanvas = querySelector("#"+data);
  double finalHeight = sourceCanvas.height * (canvas.width / sourceCanvas.width);
  context.drawImageScaled(sourceCanvas, 0, (canvas.height - powerHeight), canvas.width, finalHeight);
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
    // var parsedListe = [
    //   {
    //     "ID":"PY-S38/005R",
    //     "Translation":"[C] Your other Character in the Front Row Center Slot gains +500 Power.\n[S] \u003cb\u003eBRAINSTORM\u003c/b\u003e [(1) Rest 2 of your Characters] Flip over the top 4 cards of your Library and put them in the Waiting Room. For each Climax card revealed this way, search your Library for up to 1 ::Puyo:: Character, reveal it, put it in your hand, and shuffle your Library. "
    //   },
    //   {
    //     "ID":"PY-S38/005R",
    //     "Translation":"[C] Your other Character in the Front Row Center Slot gains +500 Power.\n[S] \u003cb\u003eBRAINSTORM\u003c/b\u003e [(1) Rest 2 of your Characters] Flip over the top 4 cards of your Library and put them in the Waiting Room. For each Climax card revealed this way, search your Library for up to 1 ::Puyo:: Character, reveal it, put it in your hand, and shuffle your Library. "
    //   },
    //   {"ID":"PY-S38/T18","Translation":"\u003cb\u003eBRAINSTORM\u003c/b\u003e Choose a Character in your Waiting Room and return it to your hand. Flip over the top 3 cards of your Library and put them in your Waiting Room. If at least 1 Climax card was revealed this way, choose a Character in your Waiting Room and return it to your hand. "}
    // ];
    // initCanvas(parsedListe);

    // HttpRequest.postFormData("/translationimages", {"url": url.value}).then((HttpRequest response) {
    //   List parsedList = JSON.decode(response.response);
    //   initCanvas(parsedList);
    // });
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

checkReturnLine (String text) {

}

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

initCanvas(List<Object> cards) {
  var xmlSeria = new XmlSerializer();
  for (var card in cards) {
    // var cardJson = JSON.decode(card);
    List<String> lines = getLines(card['Translation']);
    var height = 20;
    var cardJson = card;
    CanvasElement canvas = new CanvasElement();

    svg.SvgElement svgEl = new svg.SvgElement.tag("svg")
    ..id = cardJson['ID']
    ..classes.add("no-print");

    for (var line in lines) {
      svg.TextElement tspan = new svg.TextElement()
        ..text = line;
      tspan.attributes = {
        'y' : height.toString(),
        'x': "10",
        'font-size': "12",
        "fill":"black"
      };
      height = height + 15;
      svgEl.append(tspan);
    }
    svgEl.attributes = {
      "width": "260",
      "height": height.toString(),
      "style": "background-color: white;"
    };

    var svsString = xmlSeria.serializeToString(svgEl);
    var blop = new Blob([svsString], "image/svg+xml;charset=utf-8");
    var url = Url.createObjectUrl(blop);
    ImageElement img = new ImageElement()
      ..src = url;


    img.onLoad.listen((e) {
      canvas
        ..width = img.width
        ..height = img.height
        ..draggable = true
        ..id = card["ID"].replaceFirst("/", "")
        ..onDragStart.listen((e) => drag(e));
      CanvasRenderingContext2D context = canvas.getContext('2d');
      context.drawImage(img, 0, 0);
    });
    querySelector("#output").append(canvas);
  }
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
      querySelector(".translation-canvas").classes.toggle("hide");
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
          ..classes.add("translation-canvas")
          ..classes.add("hide")
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

printTranslation() {
  InputElement deckUrl = querySelector("#url");

  HttpRequest.postFormData("/translationimages", {"url": deckUrl.value}).then((HttpRequest response) {
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
  });
}

void main() {
  querySelector("#input-file").onChange.listen((e) => getPdfFile());
  querySelector("#send-url").onClick.listen((e) => getCardImages());
  querySelector("#print-translation").onClick.listen((e) => printTranslation());
  querySelector("#toggle-hide-translation").onClick.listen((e) => toggleHideTranslation());
  // querySelector("#yay").onDragStart.listen((e) => drag(e));

  // CanvasElement canvas = querySelector('#myCanvas');
  // ImageElement img = new ImageElement(src: "./love.png");
  // querySelector("body").append(img);
  // img.onLoad.listen((e) => initCanvas(img, canvas));

//   querySelector("#enterUrl").onclick.listen((event) {});
}
