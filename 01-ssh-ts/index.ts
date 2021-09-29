import * as pulumi from "@pulumi/pulumi";
import * as linode from "@pulumi/linode";
import { readFileSync } from 'fs';

export type Instance = {
  name: pulumi.Output<string>;
  ip: pulumi.Output<string>;
};

export let instances: Instance[] = [];

const keys = readFileSync('../../keys/2021h2/authorized_keys.'.concat(pulumi.getStack()), 'utf-8');

keys.split('\n').forEach(function (key) {
  if ( key != '' ) {
    const student = key.split(' ')[2].split('@')[0];

    // Create a Linode resource (Linode Instance)
    const instance = new linode.Instance(student, {
        authorizedKeys: [key],
        label: student,
        type: "g6-nanode-1",
        region: "eu-central",
        image: "linode/ubuntu20.04",
    });

    const instance_output: Instance = {
      name: instance.label,
      ip: instance.ipAddress,
    };
    instances.push(instance_output);
  }
})
