#!/usr/bin/env node

const {
  BaseJavaCstVisitor,
  BaseJavaCstVisitorWithDefaults
} = require("java-parser");
const { parse } = require("java-parser");
const fs = require("fs")

var args = process.argv.slice(2);
let javaText = fs.readFileSync(args[0], {encoding: 'utf8'});

const cst = parse(javaText);

const extract = (loc) => {
    return javaText.substring(loc.startOffset, loc.endOffset+1)
}

// Use "BaseJavaCstVisitor" if you need to implement all the visitor methods yourself.
class FieldExtractor extends BaseJavaCstVisitorWithDefaults {
  constructor() {
    super();
    this.customResult = [];
    this.validateVisitor();
  }
  fieldDeclaration(ctx, param) {
      //console.log(ctx.variableDeclaratorList[0].children.variableDeclarator[0].children.variableDeclaratorId[0].children.Identifier, param)
      let loc = ctx.variableDeclaratorList[0].children.variableDeclarator[0].children.variableDeclaratorId[0].location;
      console.log(extract(ctx.unannType[0].location) + "|"
          + extract(loc) + "|"
          + loc.startLine)

  }
  typeIdentifier(ctx, param) {
      console.log(ctx.Identifier[0].image)
  }
}

const fieldExtractor = new FieldExtractor();
// The CST result from the previous code snippet
fieldExtractor.visit(cst);
