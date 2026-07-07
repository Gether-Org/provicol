#!/usr/bin/env bash

set -e

LAST_TAG=$(git tag --sort=-v:refname | head -n 1)
if [ -z "$LAST_TAG" ]; then
    echo "No tag, using default v0.0.0"
    LAST_TAG="v0.0.0"
fi
VERSION_NUMBERS="${LAST_TAG#v}"
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION_NUMBERS"

MODE=${1:-patch}
case "$MODE" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Erreur : Unknown arg '$MODE'. Please use 'major', 'minor' or 'patch'."
        exit 1
        ;;
esac

NEXT_TAG="v${MAJOR}.${MINOR}.${PATCH}"

echo -n "Are you sure you want to push as $NEXT_TAG ? (y/n) : "
read -r ANSWER
if [ "$ANSWER" != "y" ] && [ "$ANSWER" != "Y" ] && [ "$ANSWER" != "yes" ]; then
    echo "Aborted"
    exit 0
fi

git tag "$NEXT_TAG"
git push origin "$NEXT_TAG"

echo "Success ! $NEXT_TAG deployed."
