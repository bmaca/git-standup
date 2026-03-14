# git-standup

`git-standup` is a tool that summarizes recent GIT activity across one ore more repositories. This is fo to answer what did I do this week? Also for me, I like to use it as a copy/paste solution for standups or weekly reports. 

## Installation

```bash
git clone https://github.com/bmaca/git-standup.git
cd git-standup
go build -o git-standup .
```

## Usages

###  What did I do last week

```
git-standup \
  -repo ~/repos/project1 \
  -repo ~/repos/project2 \
  -author "bmaca" \
  -last-week \
  -markdown
```
###  What did I do this week

```
git-standup \
  -repo ~/repos/project1 \
  -repo ~/repos/project2 \
  -author "bmaca" \
  -this-week \
```
###  What did I do custom

```
git-standup \
  -repo ~/repos/project1 \
  -repo ~/repos/project2 \
  -author "bmaca" \
  -since 2025-01-01 \
  -until 2026
  -markdown
```
