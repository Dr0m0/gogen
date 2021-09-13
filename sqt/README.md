# Gogen parser

Includes the following functionallities:
- Compile from .yaml to .go(using GORM).
- Parser files from gogen format.

# SQT(Structured Query Template)

Is a template used to build SSG(Static-Site-Generator) and SSR(Server-Side-Rendering) that is based on SQL and uses Javascript at its core. Examples:

Use a template from another file(supports both absolute and relative path):
```
@@ PASTE /path/to/template @@
```

Query data from database and generate html format(SSG), this will update the html file every time that the used table has its values changed:
```
@@  QUERY
      SELECT
        *
      FROM
        table1
    NAMED
      table1
    QUERY
      SELECT
        *
      FROM
        table2
    NAMED
      table2
    GENERATE
      (data) => (`
        <div>
          <p>${data.table1.column}</p>
          <p>${data.table2.column}</p>
        </div>
      `)
@@
```
