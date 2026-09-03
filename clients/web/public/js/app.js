const loginView =
  document.getElementById(
    "loginView"
  );

const dashboardView =
  document.getElementById(
    "dashboardView"
  );

const loginForm =
  document.getElementById(
    "loginForm"
  );

const loginMessage =
  document.getElementById(
    "loginMessage"
  );

const userInfo =
  document.getElementById(
    "userInfo"
  );

const logoutButton =
  document.getElementById(
    "logoutButton"
  );

const contentView =
  document.getElementById(
    "contentView"
  );

let session = {
  token: null,
  userId: null,
  role: null
};

// =====================================
// AUTH
// =====================================

function showLogin() {

  loginView.classList.remove(
    "hidden"
  );

  dashboardView.classList.add(
    "hidden"
  );

  logoutButton.classList.add(
    "hidden"
  );
}

function showDashboard() {

  loginView.classList.add(
    "hidden"
  );

  dashboardView.classList.remove(
    "hidden"
  );

  logoutButton.classList.remove(
    "hidden"
  );

  userInfo.textContent =
    `${session.userId} | ${session.role}`;
}

loginForm.addEventListener(
  "submit",
  async event => {

    event.preventDefault();

    loginMessage.textContent =
      "Validando...";

    const username =
      document
        .getElementById(
          "username"
        )
        .value
        .trim();

    const password =
      document
        .getElementById(
          "password"
        )
        .value;

    try {

      const response =
        await fetch(
          "/api/auth/login",
          {
            method:
              "POST",

            headers: {
              "Content-Type":
                "application/json"
            },

            body:
              JSON.stringify({
                username,
                password
              })
          }
        );

      const data =
        await response.json();

      if (
        !response.ok ||
        !data.success
      ) {

        throw new Error(
          data.message ||
          "No fue posible iniciar sesión"
        );
      }

      session = {
        userId:
          data.user.userId,

        role:
          data.user.role,

        token:
          data.token
      };

      sessionStorage.setItem(
        "upbSession",
        JSON.stringify(
          session
        )
      );

      loginMessage.textContent =
        "";

      showDashboard();

    } catch (error) {

      loginMessage.textContent =
        error.message;
    }
  }
);

logoutButton.addEventListener(
  "click",
  async () => {

    try {

      await fetch(
        "/api/auth/logout",
        {
          method:
            "POST"
        }
      );

    } catch (error) {

      console.error(
        "LOGOUT ERROR:",
        error.message
      );
    }

    sessionStorage.removeItem(
      "upbSession"
    );

    session = {
      token: null,
      userId: null,
      role: null
    };

    contentView.innerHTML =
      "<h3>Seleccione un módulo</h3>";

    showLogin();
  }
);

// =====================================
// API HELPER
// =====================================

async function apiFetch(
  url,
  options = {}
) {

  const headers = {
    ...options.headers,
    Authorization:
      `Bearer ${session.token}`
  };

  return fetch(
    url,
    {
      ...options,
      headers
    }
  );
}

// =====================================
// PHOTOS
// =====================================

async function loadPhotos() {

  contentView.innerHTML = `
    <div class="section-header">

      <div>
        <h3>
          Mis fotos
        </h3>

        <p>
          Imágenes almacenadas en el Home.
        </p>
      </div>

      <button
        id="refreshPhotos"
      >
        Actualizar
      </button>

    </div>

    <form
      id="uploadPhotoForm"
      class="upload-form"
    >

      <label>
        Seleccionar imagen

        <input
          id="photoFile"
          name="file"
          type="file"
          accept="image/*"
          required
        >
      </label>

      <label>
        Tags

        <input
          id="photoTags"
          type="text"
          placeholder="ejemplo: proyecto, laboratorio"
        >
      </label>

      <button
        type="submit"
      >
        Subir imagen
      </button>

    </form>

    <p
      id="uploadPhotoMessage"
      class="message"
    ></p>

    <p
      id="photosMessage"
    >
      Cargando fotografías...
    </p>

    <div
      id="photoGrid"
      class="photo-grid"
    ></div>
  `;

  document
    .getElementById(
      "refreshPhotos"
    )
    .addEventListener(
      "click",
      loadPhotos
    );

  document
    .getElementById(
      "uploadPhotoForm"
    )
    .addEventListener(
      "submit",
      uploadPhoto
    );

  const message =
    document.getElementById(
      "photosMessage"
    );

  const grid =
    document.getElementById(
      "photoGrid"
    );

  try {

    const response =
      await apiFetch(
        "/api/photos"
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible cargar las fotografías"
      );
    }

    const photos =
      data.photos ||
      [];

    if (
      photos.length === 0
    ) {

      message.textContent =
        "No hay fotografías almacenadas.";

      return;
    }

    message.textContent =
      `${photos.length} fotografía(s) encontrada(s).`;

    for (
      const photo of photos
    ) {

      const card =
        document.createElement(
          "article"
        );

      card.className =
        "photo-card";

      const preview =
        document.createElement(
          "div"
        );

      preview.className =
        "photo-preview";

      const image =
        document.createElement(
          "img"
        );

      image.alt =
        photo.name ||
        "Fotografía";

      image.loading =
        "lazy";

      preview.appendChild(
        image
      );

      const info =
        document.createElement(
          "div"
        );

      info.className =
        "photo-info";

      const tags =
        Array.isArray(
          photo.tags
        )
          ? photo.tags
          : [];

      info.innerHTML = `
        <strong>
          ${escapeHtml(
            photo.name ||
            "Sin nombre"
          )}
        </strong>

        <small>
          ${escapeHtml(
            photo.format ||
            photo.mimeType ||
            ""
          )}
        </small>

        <small>
          ${formatBytes(
            Number(
              photo.size ||
              0
            )
          )}
        </small>

        <div class="tags">
          ${
            tags
              .map(
                tag =>
                  `<span>${escapeHtml(tag)}</span>`
              )
              .join("")
          }
        </div>

        <button
          class="delete-photo-button"
          data-photo-id="${escapeHtml(photo.id)}"
          data-photo-name="${escapeHtml(photo.name || "Sin nombre")}"
        >
          Eliminar
        </button>
      `;

      card.appendChild(
        preview
      );

      card.appendChild(
        info
      );

      grid.appendChild(
        card
      );

      const deleteButton =
        card.querySelector(
          ".delete-photo-button"
        );

      deleteButton.addEventListener(
        "click",
        () => {
          deletePhoto(
            photo.id,
            photo.name
          );
        }
      );

      loadProtectedImage(
        image,
        photo.id
      );
    }

  } catch (error) {

    message.textContent =
      error.message;
  }
}

async function deletePhoto(
  photoId,
  photoName
) {

  const confirmed =
    window.confirm(
      `¿Eliminar la imagen "${photoName}"?`
    );

  if (!confirmed) {
    return;
  }

  try {

    const response =
      await apiFetch(
        `/api/photos/${photoId}`,
        {
          method:
            "DELETE"
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible eliminar la imagen"
      );
    }

    await loadPhotos();

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

// =====================================
// UPLOAD PHOTO
// =====================================

async function uploadPhoto(
  event
) {

  event.preventDefault();

  const fileInput =
    document.getElementById(
      "photoFile"
    );

  const tagsInput =
    document.getElementById(
      "photoTags"
    );

  const message =
    document.getElementById(
      "uploadPhotoMessage"
    );

  const file =
    fileInput.files[0];

  if (!file) {

    message.textContent =
      "Debe seleccionar una imagen.";

    return;
  }

  const formData =
    new FormData();

  formData.append(
    "file",
    file
  );

  const tags =
    tagsInput.value.trim();

  if (tags) {

    formData.append(
      "tags",
      tags
    );
  }

  message.textContent =
    "Subiendo imagen...";

  try {

    const response =
      await apiFetch(
        "/api/photos",
        {
          method:
            "POST",

          body:
            formData
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible subir la imagen"
      );
    }

    message.textContent =
      "Imagen subida correctamente.";

    fileInput.value =
      "";

    tagsInput.value =
      "";

    await loadPhotos();

  } catch (error) {

    message.textContent =
      error.message;
  }
}

// =====================================
// PROTECTED IMAGE
// =====================================

async function loadProtectedImage(
  imageElement,
  photoId
) {

  try {

    const response =
      await apiFetch(
        `/api/photos/${photoId}/content`
      );

    if (!response.ok) {

      throw new Error(
        "No se pudo obtener la imagen"
      );
    }

    const blob =
      await response.blob();

    const objectUrl =
      URL.createObjectURL(
        blob
      );

    imageElement.src =
      objectUrl;

    imageElement.onload =
      () => {

        URL.revokeObjectURL(
          objectUrl
        );
      };

  } catch (error) {

    console.error(
      "Error cargando imagen:",
      error
    );

    imageElement.alt =
      "Imagen no disponible";
  }
}

// =====================================
// FILES
// =====================================

async function loadFiles() {

  contentView.innerHTML = `
    <div class="section-header">

      <div>
        <h3>
          Mis archivos
        </h3>

        <p>
          Archivos propios y compartidos
          disponibles en el Home.
        </p>
      </div>

      <button
        id="refreshFiles"
      >
        Actualizar
      </button>

    </div>

    <form
      id="uploadFileForm"
      class="upload-form"
    >

      <label>
        Seleccionar archivo

        <input
          id="sharedFileInput"
          type="file"
          required
        >
      </label>

      <label>
        Directorio relativo

        <input
          id="relativeDirectory"
          type="text"
          placeholder="ejemplo: documentos/proyecto"
        >
      </label>

      <button
        type="submit"
      >
        Subir archivo
      </button>

    </form>

    <section
      id="storagePanel"
      class="storage-panel"
    >
      <div class="storage-header">

        <div>
          <strong>
            Almacenamiento del Home
          </strong>

          <small
            id="storagePath"
          >
            Cargando...
          </small>
        </div>

        <div
          id="storageSummary"
          class="storage-summary"
        >
          ...
        </div>

      </div>

      <div
        class="storage-bar"
      >
        <div
          id="storageProgress"
          class="storage-progress"
        ></div>
      </div>

      <small
        id="storagePercentage"
      >
        0 %
      </small>
    </section>

    <p
      id="fileMessage"
      class="message"
    >
      Cargando archivos...
    </p>

    <div
      id="fileGrid"
      class="file-grid"
    ></div>
  `;

  document
    .getElementById(
      "refreshFiles"
    )
    .addEventListener(
      "click",
      loadFiles
    );

  document
    .getElementById(
      "uploadFileForm"
    )
    .addEventListener(
      "submit",
      uploadSharedFile
    );

  await loadHomeUsage();

  const message =
    document.getElementById(
      "fileMessage"
    );

  const grid =
    document.getElementById(
      "fileGrid"
    );

  try {

    const response =
      await apiFetch(
        `/api/files?userId=${encodeURIComponent(
          session.userId
        )}`
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible cargar los archivos"
      );
    }

    const files =
      data.files ||
      [];

    if (
      files.length === 0
    ) {

      message.textContent =
        "No hay archivos disponibles.";

      return;
    }

    message.textContent =
      `${files.length} archivo(s) encontrado(s).`;

    for (
      const file of files
    ) {

      const card =
        document.createElement(
          "article"
        );

      card.className =
        "file-card";

      const isOwner =
        file.ownerId ===
        session.userId;

      card.innerHTML = `
        <div class="file-card-main">

          <strong>
            ${escapeHtml(
              file.name ||
              "Sin nombre"
            )}
          </strong>

          <small>
            ${escapeHtml(
              file.type ||
              "application/octet-stream"
            )}
          </small>

          <small>
            ${formatBytes(
              Number(
                file.size ||
                0
              )
            )}
          </small>

          <small>
            Propietario:
            ${escapeHtml(
              file.ownerId ||
              ""
            )}
          </small>

          <small>
            ${
              isOwner
                ? "Archivo propio"
                : "Compartido conmigo"
            }
          </small>

        </div>

        <div class="file-actions">

          <button
            class="download-file-button"
          >
            Descargar
          </button>

          <button
            class="update-file-button"
          >
            Actualizar contenido
          </button>

          <button
            class="share-file-button"
          >
            Compartir
          </button>

          ${
            isOwner
              ? `
                <button
                  class="revoke-file-button"
                >
                  Revocar acceso
                </button>

                <button
                  class="delete-file-button"
                >
                  Eliminar
                </button>
              `
              : ""
          }

        </div>
      `;

      card
        .querySelector(
          ".download-file-button"
        )
        .addEventListener(
          "click",
          () => {
            downloadSharedFile(
              file.id,
              file.name
            );
          }
        );

      card
        .querySelector(
          ".update-file-button"
        )
        .addEventListener(
          "click",
          () => {
            updateSharedFile(
              file.id,
              file.name
            );
          }
        );

      card
        .querySelector(
          ".share-file-button"
        )
        .addEventListener(
          "click",
          () => {
            shareSharedFile(
              file.id,
              file.name
            );
          }
        );

      if (isOwner) {

        card
          .querySelector(
            ".delete-file-button"
          )
          .addEventListener(
            "click",
            () => {
              deleteSharedFile(
                file.id,
                file.name
              );
            }
          );

        card
          .querySelector(
            ".revoke-file-button"
          )
          .addEventListener(
            "click",
            () => {
              revokeSharedFile(
                file.id,
                file.name
              );
            }
          );
      }

      grid.appendChild(
        card
      );
    }

  } catch (error) {

    message.textContent =
      error.message;
  }
}

async function loadHomeUsage() {

  const pathElement =
    document.getElementById(
      "storagePath"
    );

  const summaryElement =
    document.getElementById(
      "storageSummary"
    );

  const progressElement =
    document.getElementById(
      "storageProgress"
    );

  const percentageElement =
    document.getElementById(
      "storagePercentage"
    );

  if (
    !pathElement ||
    !summaryElement ||
    !progressElement ||
    !percentageElement
  ) {

    return;
  }

  try {

    const response =
      await apiFetch(
        `/api/home?userId=${encodeURIComponent(
          session.userId
        )}`
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success ||
      !data.home
    ) {

      throw new Error(
        data.message ||
        "No fue posible consultar el Home"
      );
    }

    const home =
      data.home;

    const quota =
      Number(
        home.quotaBytes ||
        0
      );

    const used =
      Number(
        home.usedBytes ||
        0
      );

    const available =
      Number(
        home.availableBytes ||
        0
      );

    const percentage =
      Number(
        home.percentage ||
        0
      );

    pathElement.textContent =
      home.basePath ||
      session.userId;

    summaryElement.textContent =
      `${formatBytes(used)} usados de ${formatBytes(quota)} · ${formatBytes(available)} disponibles`;

    const visualPercentage =
      Math.min(
        Math.max(
          percentage,
          0
        ),
        100
      );

    progressElement.style.width =
      `${visualPercentage}%`;

    percentageElement.textContent =
      `${percentage.toFixed(4)} % utilizado`;

  } catch (error) {

    pathElement.textContent =
      "Home no disponible";

    summaryElement.textContent =
      error.message;

    progressElement.style.width =
      "0%";

    percentageElement.textContent =
      "";
  }
}

async function uploadSharedFile(
  event
) {

  event.preventDefault();

  const input =
    document.getElementById(
      "sharedFileInput"
    );

  const directoryInput =
    document.getElementById(
      "relativeDirectory"
    );

  const message =
    document.getElementById(
      "fileMessage"
    );

  const file =
    input.files[0];

  if (!file) {

    message.textContent =
      "Debe seleccionar un archivo.";

    return;
  }

  const form =
    new FormData();

  form.append(
    "file",
    file
  );

  form.append(
    "userId",
    session.userId
  );

  const relativeDirectory =
    directoryInput.value.trim();

  if (relativeDirectory) {

    form.append(
      "relativeDirectory",
      relativeDirectory
    );
  }

  message.textContent =
    "Subiendo archivo...";

  try {

    const response =
      await apiFetch(
        "/api/files",
        {
          method:
            "POST",

          body:
            form
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible subir el archivo"
      );
    }

    input.value =
      "";

    directoryInput.value =
      "";

    await loadFiles();

  } catch (error) {

    message.textContent =
      error.message;
  }
}

async function downloadSharedFile(
  fileId,
  fileName
) {

  try {

    const response =
      await apiFetch(
        `/api/files/${fileId}/content`
      );

    if (!response.ok) {

      const data =
        await response.json();

      throw new Error(
        data.message ||
        "No fue posible descargar el archivo"
      );
    }

    const blob =
      await response.blob();

    const url =
      URL.createObjectURL(
        blob
      );

    const link =
      document.createElement(
        "a"
      );

    link.href =
      url;

    link.download =
      fileName ||
      "archivo";

    document.body.appendChild(
      link
    );

    link.click();

    link.remove();

    URL.revokeObjectURL(
      url
    );

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

async function updateSharedFile(
  fileId,
  fileName
) {

  const input =
    document.createElement(
      "input"
    );

  input.type =
    "file";

  input.style.display =
    "none";

  document.body.appendChild(
    input
  );

  input.addEventListener(
    "change",
    async () => {

      const file =
        input.files[0];

      if (!file) {

        input.remove();

        return;
      }

      const confirmed =
        window.confirm(
          `¿Reemplazar el contenido de "${fileName}" con el archivo seleccionado?`
        );

      if (!confirmed) {

        input.remove();

        return;
      }

      const form =
        new FormData();

      form.append(
        "file",
        file
      );

      try {

        const response =
          await apiFetch(
            `/api/files/${fileId}`,
            {
              method:
                "PUT",

              body:
                form
            }
          );

        const data =
          await response.json();

        if (
          !response.ok ||
          !data.success
        ) {

          throw new Error(
            data.message ||
            "No fue posible actualizar el archivo"
          );
        }

        window.alert(
          "Contenido actualizado correctamente."
        );

        await loadFiles();

      } catch (error) {

        window.alert(
          error.message
        );
      }

      input.remove();
    }
  );

  input.click();
}

// =====================================
// DELETE SHARED FILE
// =====================================

async function deleteSharedFile(
  fileId,
  fileName
) {

  const confirmed =
    window.confirm(
      `¿Eliminar el archivo "${fileName}"?`
    );

  if (!confirmed) {
    return;
  }

  try {

    const response =
      await apiFetch(
        `/api/files/${fileId}`,
        {
          method:
            "DELETE"
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible eliminar el archivo"
      );
    }

    await loadFiles();

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

async function revokeSharedFile(
  fileId,
  fileName
) {

  const targetUserId =
    window.prompt(
      `Revocar acceso a "${fileName}" para usuario:`,
      "user-002"
    );

  if (!targetUserId) {
    return;
  }

  const confirmed =
    window.confirm(
      `¿Revocar acceso de ${targetUserId.trim()} a "${fileName}"?`
    );

  if (!confirmed) {
    return;
  }

  try {

    const response =
      await apiFetch(
        `/api/files/${fileId}/share/${encodeURIComponent(
          targetUserId.trim()
        )}`,
        {
          method:
            "DELETE"
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible revocar el acceso"
      );
    }

    window.alert(
      "Acceso revocado correctamente."
    );

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

async function shareSharedFile(
  fileId,
  fileName
) {

  const targetUserId =
    window.prompt(
      `Compartir "${fileName}" con usuario:`,
      "user-002"
    );

  if (!targetUserId) {
    return;
  }

  const canWrite =
    window.confirm(
      "¿Conceder permiso de escritura?"
    );

  const canShare =
    window.confirm(
      "¿Permitir que el usuario vuelva a compartir este archivo?"
    );

  try {

    const response =
      await apiFetch(
        `/api/files/${fileId}/share`,
        {
          method:
            "POST",

          headers: {
            "Content-Type":
              "application/json"
          },

          body:
            JSON.stringify({
              targetUserId:
                targetUserId.trim(),

              canRead:
                true,

              canWrite,

              canShare
            })
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible compartir el archivo"
      );
    }

    window.alert(
      "Archivo compartido correctamente."
    );

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

// =====================================
// STREAMING
// =====================================

async function loadVideos() {

  contentView.innerHTML = `
    <div class="section-header">

      <div>
        <h3>
          Streaming
        </h3>

        <p>
          Videos del Home disponibles según tus permisos.
        </p>
      </div>

      <button
        id="refreshVideos"
      >
        Actualizar
      </button>

    </div>

    <p
      id="videosMessage"
    >
      Cargando videos...
    </p>

    <div
      id="videoGrid"
      class="photo-grid"
    ></div>
  `;

  document
    .getElementById(
      "refreshVideos"
    )
    .addEventListener(
      "click",
      loadVideos
    );

  const message =
    document.getElementById(
      "videosMessage"
    );

  const grid =
    document.getElementById(
      "videoGrid"
    );

  try {

    const response =
      await apiFetch(
        "/api/videos"
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible cargar los videos"
      );
    }

    const videos =
      data.videos ||
      [];

    if (
      videos.length === 0
    ) {

      message.textContent =
        "No hay videos autorizados.";

      return;
    }

    message.textContent =
      `${videos.length} video(s) disponible(s).`;

    for (
      const video of videos
    ) {

      const card =
        document.createElement(
          "article"
        );

      card.className =
        "photo-card";

      card.innerHTML = `
        <div
          class="photo-preview"
        >
          <video
            controls
            preload="metadata"
            playsinline
            style="width:100%;max-height:320px;background:#000"
            src="/api/stream/${encodeURIComponent(
              video.idVideo
            )}"
          ></video>
        </div>

        <div
          class="photo-info"
        >
          <strong>
            ${escapeHtml(
              video.nombre ||
              "Video"
            )}
          </strong>

          <p>
            Duración:
            ${Number(
              video.duracionSegundos ||
              0
            )} s
          </p>

          <p>
            Calidad:
            ${escapeHtml(
              video.calidad ||
              "N/D"
            )}
          </p>

          <p>
            Formato:
            ${escapeHtml(
              video.formato ||
              "N/D"
            )}
          </p>

          <p>
            Tamaño:
            ${formatBytes(
              Number(
                video.tamano ||
                0
              )
            )}
          </p>
        </div>
      `;

      grid.appendChild(
        card
      );
    }

  } catch (error) {

    message.textContent =
      error.message;
  }
}

// =====================================
// ALBUMS
// =====================================

async function loadAlbums() {

  contentView.innerHTML = `
    <div class="section-header">

      <div>
        <h3>
          Mis álbumes
        </h3>

        <p>
          Organice sus fotografías
          en colecciones.
        </p>
      </div>

      <button
        id="refreshAlbums"
      >
        Actualizar
      </button>

    </div>

    <form
      id="createAlbumForm"
      class="upload-form"
    >

      <label>
        Nombre del álbum

        <input
          id="albumName"
          type="text"
          maxlength="150"
          required
        >
      </label>

      <label>
        Descripción

        <input
          id="albumDescription"
          type="text"
          placeholder="Descripción opcional"
        >
      </label>

      <button
        type="submit"
      >
        Crear álbum
      </button>

    </form>

    <p
      id="albumMessage"
      class="message"
    >
      Cargando álbumes...
    </p>

    <div
      id="albumGrid"
      class="album-grid"
    ></div>
  `;

  document
    .getElementById(
      "refreshAlbums"
    )
    .addEventListener(
      "click",
      loadAlbums
    );

  document
    .getElementById(
      "createAlbumForm"
    )
    .addEventListener(
      "submit",
      createAlbum
    );

  const message =
    document.getElementById(
      "albumMessage"
    );

  const grid =
    document.getElementById(
      "albumGrid"
    );

  try {

    const response =
      await apiFetch(
        "/api/albums"
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible cargar los álbumes"
      );
    }

    const albums =
      data.albums ||
      [];

    if (
      albums.length === 0
    ) {

      message.textContent =
        "No hay álbumes creados.";

      return;
    }

    message.textContent =
      `${albums.length} álbum(es) encontrado(s).`;

    for (
      const album of albums
    ) {

      const card =
        document.createElement(
          "article"
        );

      card.className =
        "album-card";

      const photoIds =
        Array.isArray(
          album.photoIds
        )
          ? album.photoIds
          : [];

      card.innerHTML = `
        <h4>
          ${escapeHtml(
            album.name ||
            "Álbum sin nombre"
          )}
        </h4>

        <p>
          ${escapeHtml(
            album.description ||
            "Sin descripción"
          )}
        </p>

        <small>
          ${photoIds.length}
          fotografía(s)
        </small>

        <button
          class="open-album-button"
        >
          Ver álbum
        </button>
      `;

      card
        .querySelector(
          ".open-album-button"
        )
        .addEventListener(
          "click",
          () => {
            openAlbum(
              album.id
            );
          }
        );

      grid.appendChild(
        card
      );
    }

  } catch (error) {

    message.textContent =
      error.message;
  }
}

async function createAlbum(
  event
) {

  event.preventDefault();

  const name =
    document
      .getElementById(
        "albumName"
      )
      .value
      .trim();

  const description =
    document
      .getElementById(
        "albumDescription"
      )
      .value
      .trim();

  const message =
    document.getElementById(
      "albumMessage"
    );

  if (!name) {

    message.textContent =
      "El nombre del álbum es obligatorio.";

    return;
  }

  message.textContent =
    "Creando álbum...";

  try {

    const response =
      await apiFetch(
        "/api/albums",
        {
          method:
            "POST",

          headers: {
            "Content-Type":
              "application/json"
          },

          body:
            JSON.stringify({
              name,
              description
            })
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible crear el álbum"
      );
    }

    await loadAlbums();

  } catch (error) {

    message.textContent =
      error.message;
  }
}

async function openAlbum(
  albumId
) {

  try {

    const [
      albumResponse,
      photosResponse
    ] =
      await Promise.all([
        apiFetch(
          `/api/albums/${albumId}`
        ),

        apiFetch(
          "/api/photos"
        )
      ]);

    const albumData =
      await albumResponse.json();

    const photosData =
      await photosResponse.json();

    if (
      !albumResponse.ok ||
      !albumData.success
    ) {

      throw new Error(
        albumData.message ||
        "No fue posible abrir el álbum"
      );
    }

    if (
      !photosResponse.ok ||
      !photosData.success
    ) {

      throw new Error(
        photosData.message ||
        "No fue posible consultar las fotografías"
      );
    }

    const album =
      albumData.album;

    const photos =
      photosData.photos ||
      [];

    const photoIds =
      Array.isArray(
        album.photoIds
      )
        ? album.photoIds
        : [];

    contentView.innerHTML = `
      <div class="section-header">

        <div>
          <h3>
            ${escapeHtml(
              album.name ||
              "Álbum"
            )}
          </h3>

          <p>
            ${escapeHtml(
              album.description ||
              "Sin descripción"
            )}
          </p>
        </div>

        <div>
          <button
            id="backToAlbums"
          >
            Volver
          </button>

          <button
            id="deleteAlbumButton"
            class="delete-photo-button"
          >
            Eliminar álbum
          </button>
        </div>

      </div>

      <p>
        ${photoIds.length}
        fotografía(s) en este álbum.
      </p>

      <h4>
        Fotografías disponibles
      </h4>

      <div
        id="albumManageGrid"
        class="photo-grid"
      ></div>
    `;

    document
      .getElementById(
        "backToAlbums"
      )
      .addEventListener(
        "click",
        loadAlbums
      );

    document
      .getElementById(
        "deleteAlbumButton"
      )
      .addEventListener(
        "click",
        () => {
          deleteAlbum(
            album.id,
            album.name
          );
        }
      );

    const grid =
      document.getElementById(
        "albumManageGrid"
      );

    if (
      photos.length === 0
    ) {

      grid.innerHTML =
        "<p>No hay fotografías disponibles.</p>";

      return;
    }

    for (
      const photo of photos
    ) {

      const isInAlbum =
        photoIds.includes(
          photo.id
        );

      const card =
        document.createElement(
          "article"
        );

      card.className =
        "photo-card";

      const preview =
        document.createElement(
          "div"
        );

      preview.className =
        "photo-preview";

      const image =
        document.createElement(
          "img"
        );

      image.alt =
        photo.name ||
        "Fotografía";

      preview.appendChild(
        image
      );

      const info =
        document.createElement(
          "div"
        );

      info.className =
        "photo-info";

      info.innerHTML = `
        <strong>
          ${escapeHtml(
            photo.name ||
            "Sin nombre"
          )}
        </strong>

        <small>
          ${
            isInAlbum
              ? "Incluida en álbum"
              : "No incluida"
          }
        </small>

        <button
          class="album-photo-action"
        >
          ${
            isInAlbum
              ? "Quitar del álbum"
              : "Agregar al álbum"
          }
        </button>
      `;

      const actionButton =
        info.querySelector(
          ".album-photo-action"
        );

      actionButton.addEventListener(
        "click",
        async () => {

          if (isInAlbum) {

            await removePhotoFromAlbum(
              album.id,
              photo.id
            );

          } else {

            await addPhotoToAlbum(
              album.id,
              photo.id
            );
          }
        }
      );

      card.appendChild(
        preview
      );

      card.appendChild(
        info
      );

      grid.appendChild(
        card
      );

      loadProtectedImage(
        image,
        photo.id
      );
    }

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

async function deleteAlbum(
  albumId,
  albumName
) {

  const confirmed =
    window.confirm(
      `¿Eliminar el álbum "${albumName}"?`
    );

  if (!confirmed) {
    return;
  }

  try {

    const response =
      await apiFetch(
        `/api/albums/${albumId}`,
        {
          method:
            "DELETE"
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible eliminar el álbum"
      );
    }

    await loadAlbums();

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

// =====================================
// ADD PHOTO TO ALBUM
// =====================================

async function addPhotoToAlbum(
  albumId,
  photoId
) {

  try {

    const response =
      await apiFetch(
        `/api/albums/${albumId}/photos/${photoId}`,
        {
          method:
            "POST"
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible agregar la fotografía"
      );
    }

    await openAlbum(
      albumId
    );

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

async function removePhotoFromAlbum(
  albumId,
  photoId
) {

  try {

    const response =
      await apiFetch(
        `/api/albums/${albumId}/photos/${photoId}`,
        {
          method:
            "DELETE"
        }
      );

    const data =
      await response.json();

    if (
      !response.ok ||
      !data.success
    ) {

      throw new Error(
        data.message ||
        "No fue posible quitar la fotografía"
      );
    }

    await openAlbum(
      albumId
    );

  } catch (error) {

    window.alert(
      error.message
    );
  }
}

// =====================================
// UTILS
// =====================================

function formatBytes(
  bytes
) {

  if (
    !Number.isFinite(bytes) ||
    bytes <= 0
  ) {

    return "0 B";
  }

  const units = [
    "B",
    "KB",
    "MB",
    "GB"
  ];

  let index =
    0;

  let value =
    bytes;

  while (
    value >= 1024 &&
    index <
      units.length - 1
  ) {

    value /=
      1024;

    index++;
  }

  return `${value.toFixed(
    index === 0
      ? 0
      : 2
  )} ${units[index]}`;
}

function escapeHtml(
  value
) {

  return String(
    value
  )
    .replaceAll(
      "&",
      "&amp;"
    )
    .replaceAll(
      "<",
      "&lt;"
    )
    .replaceAll(
      ">",
      "&gt;"
    )
    .replaceAll(
      '"',
      "&quot;"
    )
    .replaceAll(
      "'",
      "&#039;"
    );
}

// =====================================
// MODULE BUTTONS
// =====================================

document
  .querySelectorAll(
    ".module-card"
  )
  .forEach(
    button => {

      button.addEventListener(
        "click",
        () => {

          const module =
            button.dataset.module;

          if (
            module ===
            "photos"
          ) {

            loadPhotos();

            return;
          }

          if (
            module ===
            "files"
          ) {

            loadFiles();

            return;
          }

          if (
            module ===
            "albums"
          ) {

            loadAlbums();

            return;
          }

          if (
            module ===
            "streaming"
          ) {

            loadVideos();

            return;
          }
        }
      );
    }
  );

// =====================================
// RESTORE SESSION
// =====================================

const stored =
  sessionStorage.getItem(
    "upbSession"
  );

if (stored) {

  try {

    session =
      JSON.parse(
        stored
      );

    showDashboard();

  } catch {

    sessionStorage.removeItem(
      "upbSession"
    );

    showLogin();
  }

} else {

  showLogin();
}
