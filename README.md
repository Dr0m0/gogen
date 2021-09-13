# Gogen

It's an application generator, I will try to make this to generate the most possible parts of an application.

**WARNING:** This project is incomplete.

### Static-site format

The static-site generator is inspired in an _imperative language_ like **SQL** for accessing data, and a _funcional language_ like **Javascript** for formating data.
The following operations are implemented:
\
Adding a template:
```
{{ PASTE /root/path/to/file }}

or

{{ PASTE relative/path/to/file }}

like

<html>
<style>
  {{ PASTE css/style.css }}
</style>
</html>
```

Formating data:
```
{{  FORMAT object_type FROM path/to/schemas.yaml AS object IN
    () => (`
      <!-- this is Javascript returning a template using the object -->
      <div>
        <p>${object.field}</p>
      </div>
    `)
}}

or

<ul>
{{  FORMAT apples FROM schemas.yaml ORDER BY DESC taste AS apple AND
      FORMAT bananas FROM schemas.yaml AS banana IN
      () => (`
        <li>${apple.name}</li>
        <li>${banana.name}</li>
      `)
}}
</ul>
```
\
Before runnning **gogen** is necessary to write a _genesis.yaml_ file that will contain all the site routes associated with a path to an .html that will represent that route in the web. The _genesis.yaml_ file defines too the API for the server.
\
**Genesis file specification**
```yaml
---
name: application_name
routes:
  - web:   /
    file:  pages/home/index.html

  - web:   /customer/:id
    file:  pages/customer/index.html
schemas:
  object:
    field: value
...
```
\
Running **gogen:**
```
gogen genesis.yaml
```
\
**Important:** The route path of the application will be the genesis.yaml directory.

### Technologies

- [esbuild](https://esbuild.github.io/api/#bundle)

### Optimizations

- [ ] On _templates/handler.go_ don't convert to _string_ but uses pure _[]byte_.

### Cloud provider support

- [ ] AWS.
- [ ] Azure.
- [ ] Google Cloud.

