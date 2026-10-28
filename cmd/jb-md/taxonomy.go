package main

import (
	"fmt"
	"os"
	"strings"
)

// ensureCategory 查 → 无则建 → 冲突再查；进程内缓存 name→uid。
func ensureCategory(client *APIClient, name string, dryRun bool, cache map[string]string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("分类名为空")
	}
	if uid, ok := cache[name]; ok {
		return uid, nil
	}

	// 1. 按 name 精确匹配已有分类
	uid, err := findTaxonomyUID(client, "category", name)
	if err != nil {
		return "", err
	}
	if uid != "" {
		fmt.Fprintf(os.Stderr, "分类「%s」已存在 uid=%s\n", name, uid)
		cache[name] = uid
		return uid, nil
	}

	// 2. dry-run 只警告，不创建
	if dryRun {
		fmt.Fprintf(os.Stderr, "警告: 分类「%s」不存在，正式上传时将自动创建\n", name)
		cache[name] = ""
		return "", nil
	}

	// 3. 创建；失败（含唯一约束）后再查一次
	uid, err = client.createTaxonomy("category", name)
	if err != nil {
		uid2, err2 := findTaxonomyUID(client, "category", name)
		if err2 != nil {
			return "", fmt.Errorf("创建分类「%s」失败: %v; 复查失败: %w", name, err, err2)
		}
		if uid2 == "" {
			return "", fmt.Errorf("创建分类「%s」失败且复查仍不存在: %w", name, err)
		}
		fmt.Fprintf(os.Stderr, "分类「%s」创建冲突后已找到 uid=%s\n", name, uid2)
		cache[name] = uid2
		return uid2, nil
	}
	fmt.Fprintf(os.Stderr, "分类「%s」不存在，已自动创建 uid=%s\n", name, uid)
	cache[name] = uid
	return uid, nil
}

func findTaxonomyUID(client *APIClient, kind, name string) (string, error) {
	list, err := client.taxonomyOptions(kind)
	if err != nil {
		return "", err
	}
	for _, item := range list {
		if item.Name == name {
			return item.UID, nil
		}
	}
	return "", nil
}
