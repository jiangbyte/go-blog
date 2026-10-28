<script setup lang="ts">
import AppIcon from '@/components/chrome/AppIcon.vue'
import { formatDate } from '@/utils/markdown'
import { projectProgressLabel, projectProgressMod } from '@/utils/projectProgress'

defineProps<{ project: any }>()
</script>

<template>
  <div class="project-head page-gutter">
    <div class="page-shell project-head__inner">
      <span
        class="project-head__progress"
        :class="`is-${projectProgressMod(project.progress)}`"
      >
        {{ projectProgressLabel(project.progress) }}
      </span>
      <div class="project-head__main">
        <div v-if="project.sticky" class="project-head__badges">
          <span class="badge badge--pin">置顶</span>
        </div>
        <h1 class="project-head__title">{{ project.title }}</h1>
        <p v-if="project.description" class="project-head__desc">
          {{ project.description }}
        </p>
        <div class="project-head__meta">
          <span class="pill">
            <AppIcon name="view" :size="14" />
            {{ Number(project.view_count) || 0 }}
          </span>
          <span class="pill">
            <AppIcon name="calendar" :size="14" />
            {{ formatDate(project.date) }}
          </span>
          <RouterLink
            v-for="tag in project.project_tags"
            :key="tag.uid"
            :to="`/project-tags/${tag.uid}`"
            class="pill pill--link"
          >
            <AppIcon name="tags" :size="12" />
            {{ tag.name }}
          </RouterLink>
        </div>
        <div v-if="project.site || project.git" class="project-head__actions">
          <a
            v-if="project.site"
            :href="project.site"
            target="_blank"
            rel="noopener noreferrer"
            class="action action--primary"
          >
            <AppIcon name="link" :size="14" />
            站点
          </a>
          <a
            v-if="project.git"
            :href="project.git"
            target="_blank"
            rel="noopener noreferrer"
            class="action"
          >
            <AppIcon name="github" :size="14" />
            GitHub
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.project-head {
  position: relative;
  z-index: 10;
  padding-top: calc(4.75rem + env(safe-area-inset-top, 0px));
  padding-bottom: 1.35rem;
  color: #fff;
}

.project-head__inner {
  position: relative;
  display: flex;
  align-items: flex-end;
  min-height: 0;
}

.project-head__progress {
  position: absolute;
  top: 0;
  right: 0;
  z-index: 2;
  padding: 0.2rem 0.65rem;
  border-radius: 999px;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.35;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
}

.project-head__progress.is-planned {
  background: linear-gradient(135deg, #8b7cf6 0%, #6d5ce7 100%);
}

.project-head__progress.is-active {
  background: linear-gradient(135deg, #4c8dff 0%, #2f6fed 100%);
}

.project-head__progress.is-done {
  background: linear-gradient(135deg, #3db887 0%, #2a9a6e 100%);
}

.project-head__progress.is-maintained {
  background: linear-gradient(135deg, #2eb8c9 0%, #1a96a8 100%);
}

.project-head__progress.is-paused {
  background: linear-gradient(135deg, #8b93a0 0%, #6b7280 100%);
}

.project-head__main {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  width: 100%;
  max-width: 42rem;
  min-width: 0;
  padding-right: 5.5rem;
}

.project-head__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.15rem 0.5rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.16);
  color: inherit;
  font-size: 0.75rem;
  line-height: 1.35;
}

.badge--pin {
  background: linear-gradient(135deg, #e85d4c 0%, #d4483a 100%);
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.project-head__title {
  margin: 0;
  font-size: clamp(1.45rem, 3.2vw, 1.85rem);
  font-weight: 700;
  line-height: 1.25;
}

.project-head__desc {
  margin: 0;
  max-width: 36rem;
  font-size: 0.92rem;
  line-height: 1.55;
  color: rgba(255, 255, 255, 0.92);
}

.project-head__meta,
.project-head__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  height: 1.7rem;
  padding: 0 0.65rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  font-size: 0.78rem;
  color: inherit;
  text-decoration: none;
  white-space: nowrap;
}

.pill--link:hover {
  background: rgba(255, 255, 255, 0.24);
}

.action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  min-height: 1.85rem;
  padding: 0.25rem 0.8rem;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  font-size: 0.84rem;
  text-decoration: none;
}

.action:hover {
  background: rgba(255, 255, 255, 0.22);
}

.action--primary {
  border-color: transparent;
  background: var(--milk-ink, #2c3444);
}

.action--primary:hover {
  opacity: 0.92;
  background: var(--milk-ink, #2c3444);
}

@media (min-width: 768px) {
  .project-head {
    padding-top: calc(4.5rem + env(safe-area-inset-top, 0px));
    padding-bottom: 1.6rem;
  }

  .project-head__inner {
    min-height: 8.5rem;
  }
}
</style>
