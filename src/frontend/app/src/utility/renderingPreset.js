export class RenderingPreset {
  constructor(name, params, objectPath, materialPath, texturePaths) {
    this.name = name;
    this.params = params;
    this.objectPath = objectPath;
    this.materialPath = materialPath;
    this.texturePaths = texturePaths;
  }
}
