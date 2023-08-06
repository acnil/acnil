# Acnil

Código fuente de la página web de ACNIL

## Imagenes

Tenemos un zip con las imágenes en alta resolución que no se encuentran en este repositorio para evitar que sea muy grande. Se pueden descargar [aqui](https://drive.google.com/file/d/1o7J-8wdfjL5hO6pA_I3OWRr1uxcwnXE1/view?usp=sharing). Hay un script que utiliza `./image-processing.yaml` para generar las imágenes en baja resolución para la web.

## Ejecutar en local

1. Instala dependencias
   1. [Hugo v0.114.0](https://github.com/gohugoio/hugo/releases/tag/v0.114.0)
   3. VScode
   4. Git
3. Descarga el repositorio
4. Ejecuta `hugo serve -D` desde la raiz
5. Enjoy!

## Ejecutar en codespaces

1. Abrir Codespace
   1. Hacer click en el boton verde arriba a la derecha en la página principal `Code`
   2. Seleccionar Codespaces
   3. Crear nuevo en master
   4. :warning: Es importante acordarse de cerrar el espacio al terminar, sino se agotará el tiempo de uso gratuito. (60h) Puedes configurar que el entorno se apague tras 30m de inactividad [aqui](https://github.com/settings/codespaces)
2. Abrir Hugo
   1. `ctrl+shift+p` y escribir `Run Task`(Ejecutar Tarea)
   2. Enter
   3. Aparecerá un terminal en la parte inferior
   4. Una notificación a la derecha abajo dira algo como

        ```txt
        La aplicación que se está ejecutando en el puerto 1313 está disponible.[Ver todos los puertos reenviados] (command:~remote.forwardedPorts.focus)
        ```

   5. Hacer click en `Abrir en el navegador`
3. Editar!
   1. Edita cualquier página y los cambios se verán en tiempo real en la ventana anterior

## Links

- [Debug cards](http://debug.iframely.com/?uri=https%3A%2F%2Fbigbangburgos.es%2Fjornadas%2F2022-07-09%2F)