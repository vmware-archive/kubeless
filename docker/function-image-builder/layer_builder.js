/**
 * Copyright (c) 2016-2017 Bitnami
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

'use strict';

/**
 * This script adds a new layer to a local container image in its OCI format
 * It uses the tar file given as parameter to include the files as a new layer
 * Then, it adds the layer info to the existing manifest and configuration files
 */

const crypto = require('crypto');
const fs = require('fs');
const path = require('path');

if (process.argv.length < 4) {
    console.error(`USAGE: ${process.argv.length[0]} ${process.argv.length[0]} /path/to/image/imageDir /path/to/function.tar`);
}

const imageDir = process.argv[2];
const functionFile = process.argv[3];

const functionSha256 = crypto.createHash('sha256').update(fs.readFileSync(functionFile)).digest('hex');
const functionSize = fs.statSync(functionFile).size;
fs.copyFileSync(functionFile, path.join(imageDir, `${functionSha256}.tar`));

const manifestFile = path.join(imageDir, 'manifest.json');
const manifest = JSON.parse(fs.readFileSync(manifestFile));

// Update config
const configFile = path.join(imageDir, `${manifest.config.digest.replace('sha256:', '')}.tar`);
const config = JSON.parse(fs.readFileSync(configFile));
//   Delete some properties that doesn't apply anymore
config.config.Hostname = "";
config.config.Image = "";
config.container = "";
config.container_config.Hostname = "";
config.container_config.Image = "";
//   Update new properties
config.created = new Date().toISOString();
config.history.push({
    created: new Date().toISOString(),
    comment: 'Made by Kubeless'
});
config.rootfs.diff_ids.push(`sha256:${functionSha256}`);

// Save new config
const configString = JSON.stringify(config);
const configNewHash = crypto.createHash('sha256').update(configString).digest('hex');
const configNewSize = Buffer.byteLength(configString, 'uft8');
console.log(`Writing new layer ${imageDir}/${configNewHash}.tar`);
fs.writeFileSync(`${imageDir}/${configNewHash}.tar`, configString);

// Update manifest
manifest.config.size = configNewSize;
manifest.config.digest = `sha256:${configNewHash}`;
manifest.layers.push({
    mediaType: 'application/vnd.docker.image.rootfs.diff.tar.gzip',
    size: functionSize,
    digest: `sha256:${functionSha256}`
});
fs.writeFileSync(manifestFile, JSON.stringify(manifest, null, 2));
console.log(`Updated manifest for ${imageDir}`);
