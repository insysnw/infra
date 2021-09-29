import * as pulumi from "@pulumi/pulumi";
import * as linode from "@pulumi/linode";
import { readFileSync } from 'fs';

export type Instance = {
  name: pulumi.Output<string>;
  ip: pulumi.Output<string>;
};

export let instances: Instance[] = [];

const nginxStackScript = new linode.StackScript("nginxStackScript", {
    label: "nginx",
    description: "Installs an NGINX Package",
    script: `#!/bin/bash
# <UDF name="package" label="System Package to Install" example="nginx" default="">
apt-get -q update && apt-get -q -y install $PACKAGE
`,
    images: [
        "linode/ubuntu20.04",
    ],
    revNote: "initial version",
});

const keys = readFileSync('../../keys/2021h2/authorized_keys.'.concat(pulumi.getStack()), 'utf-8');

let config = new pulumi.Config();

keys.split('\n').forEach(function (key) {
  if ( key != '' ) {
    const student = key.split(' ')[2].split('@')[0];

    // Create a Linode resource (Linode Instance)
    const instance = new linode.Instance(student, {
        authorizedKeys: [key],
        authorizedUsers: config.getObject("tutors"),
        label: student,
        privateIp: true,
        type: "g6-nanode-1",
        region: "eu-central",
        image: "linode/ubuntu20.04",
        stackscriptId: nginxStackScript.id.apply(parseInt),
        stackscriptData: {
            "package": "nginx",
        },
    }, { deleteBeforeReplace: true });

    const instance_output: Instance = {
      name: instance.label,
      ip: instance.ipAddress,
    };
    instances.push(instance_output);
  }
})
