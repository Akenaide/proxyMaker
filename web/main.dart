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

class myNodeValidator implements NodeValidator {
  bool allowsAttribute(Element element, String attributeName, String value) {
    return true;
  }

  bool allowsElement(Element element) => true;

  myNodeValidator() : super();
}

NodeValidator validator = new myNodeValidator();

void toggleHideTranslation() {
  CanvasElement canvas = querySelector(".translation-canvas");
  canvas.classes.toggle("hide");
}

addImage(event) {
  var image = new ImageElement()..src = event.target.src;

  CanvasElement canvas = new CanvasElement()
    ..classes.add("image");

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
  HttpRequest.postFormData("/views/cardimages", {"url": url.value}).then(
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

printTranslation() {
  spinner.classes.toggle("hide");
  InputElement deckUrl = querySelector("#url");

  HttpRequest.postFormData("/views/translationimages", {"url": deckUrl.value}).then(
      (HttpRequest response) {
    List parsedList = JSON.decode(response.response);
    for (var card in parsedList) {
      DivElement printDiv = new DivElement();
      printDiv.appendHtml(card["Translation"].replaceAll("\u21b5", ""),
          validator: validator);
      printDiv.classes.add("translation-block");
      for (var td in printDiv.querySelectorAll("td")) {
        td.attributes["style"] = "";
      }
      printDiv.querySelector("table").attributes["width"] = "";
      ImageElement image = printDiv.querySelector("img");
      image
        ..src = card["URL"]
        ..classes.add("mini-image");

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

  HttpRequest.postFormData("/views/estimateprice", {"url": deckUrl.value}).then(
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
      row.addCell().text = (card["Price"] * card["Amount"]).toString();
    }
    querySelector('#output').append(table);
    spinner.classes.toggle("hide");
  });
}

void main() {
  querySelector("#send-url").onClick.listen((e) => getCardImages());
  querySelector("#print-translation").onClick.listen((e) => printTranslation());
  querySelector("#estimate-price").onClick.listen((e) => estimatePrice());
  querySelector("#toggle-hide-translation")
      .onClick
      .listen((e) => toggleHideTranslation());
  spinner = querySelector("#spinner");
}
