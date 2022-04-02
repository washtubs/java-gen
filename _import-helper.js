#!/usr/bin/env node

const {
  BaseJavaCstVisitor,
  BaseJavaCstVisitorWithDefaults
} = require("java-parser");
const { parse } = require("java-parser");
const fs = require("fs")

//const javaText = `
//import java.util.Optional;

//import javax.annotation.Nonnull;
//import javax.annotation.Nullable;

//import org.apache.commons.lang3.builder.EqualsBuilder;
//import org.apache.commons.lang3.builder.HashCodeBuilder;
//import org.apache.commons.lang3.builder.ReflectionToStringBuilder;

//import org.gtri.ctisl.sis.foxhound.config.ExportToTypescript;
//import org.gtri.ctisl.sis.foxhound.services.file.FileStatuses;
//import java.util.Map;

//import static java.util.Objects.requireNonNull;
//public class HelloWorldExample{
  //private final int foo;
  //private final String bar;
  //private final Map<String, Integer> baz;
  //public static void main(String args[]){
    //final String notField = "hi";
    //System.out.println("Hello World !");
  //}
//}
//`;

var args = process.argv.slice(2);
let file = args[0];
let javaText;

if(process.stdin.isTTY) {
  javaText = fs.readFileSync(file, {encoding: 'utf8'});
} else {
  javaText = fs.readFileSync(0, {encoding: 'utf-8'});
}

const cst = parse(javaText);

const structure = require('./imports.json')

// Use "BaseJavaCstVisitor" if you need to implement all the visitor methods yourself.
class ImportsExtractor extends BaseJavaCstVisitorWithDefaults {
  constructor() {
    super();
    this.importsStart = 999;
    this.importsEnd = 0;
    this.importStatements = [];
    this.customResult = [];
    this.validateVisitor();
  }
  importDeclaration(ctx) {
      //console.log(ctx.variableDeclaratorList[0].children.variableDeclarator[0].children.variableDeclaratorId[0].children.Identifier, param)

      let loc = ctx.Import[0];
      if (this.importsStart > loc.startLine) {
          this.importsStart = loc.startLine;
      }
      if (this.importsEnd < loc.endLine) {
          this.importsEnd = loc.endLine;
      }

      let startOffset = ctx.Import[0].startOffset;
      let endOffset = ctx.Semicolon[0].endOffset;
      let statement = javaText.substring(startOffset, endOffset + 1);
      this.importStatements.push(statement);

  }
}

const importsExtractor = new ImportsExtractor();
importsExtractor.visit(cst);

let allImports = importsExtractor.importStatements;
let groupTexts = [];
for (let i = 0; i < structure.length; i++) {
    let group = structure[i];
    let groupText = ''
    for (let j = 0; j < group.length; j++) {
        let pattern = group[j];
        let filtered = allImports.filter( (statement) => statement.match(pattern) );
        let sorted = filtered.sort();
        sorted.forEach((statement) => {
            groupText += statement + "\n"
        });
    }
    groupTexts.push(groupText);
}

let newText = groupTexts.filter(g => g.length !== 0).join("\n")

console.log(JSON.stringify({
    changes: {
        ["file://" + file]: [ {
            range: {
                end: {
                    line: importsExtractor.importsEnd,
                    character: 9999
                },
                start: {
                    line: importsExtractor.importsStart-1,
                    character: 0
                }
            },
            newText: newText
        } ]
    }
}));
