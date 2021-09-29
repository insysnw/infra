import * as pulumi from "@pulumi/pulumi";
import * as linode from "@pulumi/linode";
import { readFileSync } from 'fs';

export type Instance = {
  name: pulumi.Output<string>;
  ip: pulumi.Output<string>;
};

export let instances: Instance[] = [];

const keys = fs.readFileSync('../../keys/2021h2/authorized_keys.'.concat(pulumi.getStack()), 'utf-8'):

keys.split(/[\r\n]+/).forEach(function (line) {
  student = line.split(/[\t\ ]+/)[2].split('@')[0];
  new SshKey(name: student, args: {label: student, sshKey: line});

  // Create a Linode resource (Linode Instance)
  const instance = new linode.Instance(student, {
      type: "g6-nanode-1",
      region: "eu-central",
      image: "linode/ubuntu20.04",
  });

  const instance_output: Instance = {
    name: instance.
    ipv4: instance.
  };
  instances.push(instance_output);
})



export const instanceLabel = instance.label;
