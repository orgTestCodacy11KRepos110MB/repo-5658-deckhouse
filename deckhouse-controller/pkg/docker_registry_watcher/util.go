package docker_registry_watcher

import (
	"fmt"
	"regexp"
	"strings"
)

// Kube with Docker should contain 'docker-pullable://' prefix and not 'docker://'.
// But containerd has no prefix at all.

// DockerPullableDigestRe detects if imageId field contains image digest
var DockerPullableDigestRe = regexp.MustCompile("docker-pullable://.*@sha256:[a-fA-F0-9]{64}")

// KubeDigestRe detects if imageId field sha256 digest
var KubeDigestRe = regexp.MustCompile(".*@sha256:[a-fA-F0-9]{64}")

// DockerImageDigestRe regexp extracts docker image digest from string
var DockerImageDigestRe = regexp.MustCompile("(sha256:?)?[a-fA-F0-9]{64}")

// var KubeImageIdRe = regexp.MustCompile("docker://sha256:[a-fA-F0-9]{64}")

// Поиск digest в строке.
// Учитывается специфика kubernetes — если есть префикс docker-pullable://, то в строке digest.
// Если префикс docker:// или нет префикса, то скорее всего там imageId, который нельзя
// применить для обновления, поэтому возвращается ошибка
// Пример строки с digest из kubernetes: docker-pullable://registry/repo:tag@sha256:DIGEST-HASH
func FindImageDigest(imageID string) (image string, err error) {
	if strings.Contains(imageID, "://") {
		if !DockerPullableDigestRe.MatchString(imageID) {
			err = fmt.Errorf("pod status contains image_id and not digest. Deckhouse update process not working in clusters with Docker 1.11 or earlier. imageID='%s', regex='%s'", imageID, DockerPullableDigestRe)
			return "", err
		}
	} else {
		if !KubeDigestRe.MatchString(imageID) {
			err = fmt.Errorf("pod status contains image_id and not digest. Deckhouse update process not working in clusters with Docker 1.11 or earlier. imageID='%s', regex='%s'", imageID, KubeDigestRe)
			return "", err
		}
	}

	image = DockerImageDigestRe.FindString(imageID)
	return image, nil
}

// Проверка, что строка это docker digest
func IsValidImageDigest(imageID string) bool {
	return DockerImageDigestRe.MatchString(imageID)
}
