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

function isNotEmpty(str: string) {
    return str.length > 0;
}

const keys = readFileSync('../../keys/2021h2/authorized_keys.'.concat(pulumi.getStack()), 'utf-8').split('\n').filter(isNotEmpty);

let config = new pulumi.Config();

keys.forEach(function (key) {
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
});

const tutorsInstance = new linode.Instance("tutors", {
    authorizedUsers: config.getObject("tutors"),
    label: "tutors",
    privateIp: true,
    type: "g6-nanode-1",
    region: "eu-central",
    image: "linode/ubuntu20.04",
    stackscriptId: nginxStackScript.id.apply(parseInt),
    stackscriptData: {
        "package": "nginx",
    },
}, { deleteBeforeReplace: true });

export const tutorsIP = tutorsInstance.ipAddress;

const privateInstance = new linode.Instance("private", {
    authorizedKeys: keys,
    authorizedUsers: config.getObject("tutors"),
    label: "private",
    privateIp: true,
    type: "g6-nanode-1",
    region: "eu-central",
    image: "linode/ubuntu20.04",
    stackscriptId: nginxStackScript.id.apply(parseInt),
    stackscriptData: {
        "package": "nginx",
    },
}, { deleteBeforeReplace: true });

var netmask = require('netmask');
const internalBlock = new netmask.Netmask('192.168.128.0/17');

function isPrivate(ip: string) {
    return internalBlock.contains(ip);
}
export const privateIP = privateInstance.ipv4s.apply(ips => ips.filter(isPrivate));

new linode.Firewall("turnPrivate", {
    label: "turnPrivate",
    tags: ["kek"],
    inbounds: [
        {
            label: "allow-internal",
            action: "ACCEPT",
            protocol: "TCP",
            ports: "1-65535",
            ipv4s: ["192.168.128.0/17"],
        },
    ],
    inboundPolicy: "DROP",
    outboundPolicy: "ACCEPT",
    linodes: [privateInstance.id.apply(parseInt)],
});
