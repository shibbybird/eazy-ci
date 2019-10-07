#!/bin/bash
# Related to this windows compile issue https://github.com/moby/moby/pull/40021/commits/c3a0a3744636069f43197eb18245aaae89f568e5
sed -i 's/sd, err := winio.SddlToSecurityDescriptor(sddl)/sd, err := windows.SecurityDescriptorFromString(sddl)/g' ./vendor/github.com/docker/docker/pkg/system/filesys_windows.go
sed -i 's/sa.SecurityDescriptor = uintptr(unsafe.Pointer(&sd\[0\]))/sa.SecurityDescriptor = sd/g' ./vendor/github.com/docker/docker/pkg/system/filesys_windows.go
sed -i 's/\twinio "github.com\/Microsoft\/go-winio"//g' ./vendor/github.com/docker/docker/pkg/system/filesys_windows.go
