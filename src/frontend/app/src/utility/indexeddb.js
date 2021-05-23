// From https://stackoverflow.com/a/41595124

export async function loadFromIndexedDB(storeName, id) {
  return new Promise(function (resolve, reject) {
    var dbRequest = indexedDB.open(storeName);

    dbRequest.onerror = function () {
      reject(Error("Failed request to IndexedDB"));
    };

    dbRequest.onupgradeneeded = function (event) {
      // Objectstore does not exist. Nothing to load
      event.target.transaction.abort();
      reject(Error("Not found"));
    };

    dbRequest.onsuccess = function (event) {
      var database = event.target.result;
      var transaction = database.transaction([storeName]);
      var objectStore = transaction.objectStore(storeName);
      var objectRequest = objectStore.get(id);

      objectRequest.onerror = function () {
        reject(Error("Failed object request to IndexedDB"));
      };

      objectRequest.onsuccess = function () {
        if (objectRequest.result) resolve(objectRequest.result);
        else reject(Error("IndexedDB object not found"));
      };
    };
  });
}

export async function saveToIndexedDB(storeName, object) {
  return new Promise(function (resolve, reject) {
    if (object.id === undefined) reject(Error("object has no id."));
    var dbRequest = indexedDB.open(storeName);

    dbRequest.onerror = function () {
      reject(Error("IndexedDB database error"));
    };

    dbRequest.onupgradeneeded = function (event) {
      var database = event.target.result;
      database.createObjectStore(storeName, {
        keyPath: "id",
      });
    };

    dbRequest.onsuccess = function (event) {
      var database = event.target.result;
      var transaction = database.transaction([storeName], "readwrite");
      var objectStore = transaction.objectStore(storeName);
      var objectRequest = objectStore.put(object); // Overwrite if exists

      objectRequest.onerror = function () {
        reject(Error("Error text"));
      };

      objectRequest.onsuccess = function () {
        resolve("Data saved OK");
      };
    };
  });
}
