<script setup lang="ts">
import { coverOf } from '@/utils/markdown'
import { projectProgressLabel, projectProgressMod } from '@/utils/projectProgress'
import { getSite } from '@/api'

defineProps<{ project: any; index: number }>()
const site = getSite()
</script>

<template>
  <article class="projects-card glass-card group">
    <RouterLink
      :to="`/projects/${project.uid}`"
      class="projects-card__cover"
      tabindex="-1"
      aria-hidden="true"
    >
      <img
        :src="coverOf(project.cover, site.default_covers, index)"
        :alt="project.title"
        class="projects-card__cover-img"
        loading="lazy"
      />
      <span v-if="project.sticky" class="projects-card__pin" aria-label="置顶">置顶</span>
      <span
        class="projects-card__progress"
        :class="`is-${projectProgressMod(project.progress)}`"
        :aria-label="`进度 ${projectProgressLabel(project.progress)}`"
      >
        {{ projectProgressLabel(project.progress) }}
      </span>
    </RouterLink>

    <div class="projects-card__body">
      <RouterLink :to="`/projects/${project.uid}`" class="projects-card__title">
        {{ project.title }}
      </RouterLink>

      <p class="projects-card__desc">{{ project.description || '\u00a0' }}</p>

      <ul class="projects-card__tags" aria-label="项目标签">
        <li v-for="tag in project.project_tags || []" :key="tag.uid">
          <RouterLink :to="`/project-tags/${tag.uid}`" class="projects-card__tag">
            {{ tag.name }}
          </RouterLink>
        </li>
      </ul>

      <div class="projects-card__actions">
        <RouterLink :to="`/projects/${project.uid}`" class="projects-card__action is-primary">
          详情
        </RouterLink>
        <a
          v-if="project.git"
          :href="project.git"
          target="_blank"
          rel="noopener noreferrer"
          class="projects-card__action"
        >
          仓库
        </a>
        <a
          v-if="project.site"
          :href="project.site"
          target="_blank"
          rel="noopener noreferrer"
          class="projects-card__action"
        >
          访问
        </a>
      </div>
    </div>
  </article>
</template>

<style scoped>
.projects-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
  height: 100%;
  overflow: hidden;
  transition:
    box-shadow 0.25s ease,
    transform 0.25s ease;
}

.projects-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(44, 52, 68, 0.06);
}

.projects-card__cover {
  position: relative;
  display: block;
  overflow: hidden;
  aspect-ratio: 2 / 1;
  background: var(--efu-secondbg);
}

.projects-card__pin {
  position: absolute;
  top: 0.65rem;
  left: 0.65rem;
  z-index: 2;
  padding: 0.15rem 0.55rem;
  border-radius: 999px;
  background: linear-gradient(135deg, #e85d4c 0%, #d4483a 100%);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.35;
  box-shadow: 0 2px 8px rgba(212, 72, 58, 0.35);
}

.projects-card__progress {
  position: absolute;
  top: 0.65rem;
  right: 0.65rem;
  z-index: 2;
  padding: 0.15rem 0.55rem;
  border-radius: 999px;
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.35;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.18);
}

.projects-card__progress.is-planned {
  background: linear-gradient(135deg, #8b7cf6 0%, #6d5ce7 100%);
}

.projects-card__progress.is-active {
  background: linear-gradient(135deg, #4c8dff 0%, #2f6fed 100%);
}

.projects-card__progress.is-done {
  background: linear-gradient(135deg, #3db887 0%, #2a9a6e 100%);
}

.projects-card__progress.is-maintained {
  background: linear-gradient(135deg, #2eb8c9 0%, #1a96a8 100%);
}

.projects-card__progress.is-paused {
  background: linear-gradient(135deg, #8b93a0 0%, #6b7280 100%);
}

.projects-card__cover-img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.35s ease;
}

.projects-card:hover .projects-card__cover-img {
  transform: scale(1.03);
}

.projects-card__body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.85rem 0.95rem 0.95rem;
  min-width: 0;
}

.projects-card__title {
  flex-shrink: 0;
  height: 1.35em;
  color: var(--milk-ink);
  font-size: 1.02rem;
  font-weight: 600;
  line-height: 1.35;
  text-decoration: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.projects-card__title:hover {
  opacity: 0.8;
}

.projects-card__desc {
  flex-shrink: 0;
  margin: 0;
  height: calc(1.55em * 2);
  color: var(--efu-secondtext);
  font-size: 0.86rem;
  line-height: 1.55;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.projects-card__tags {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 0.35rem;
  margin: 0;
  padding: 0;
  list-style: none;
  flex-shrink: 0;
  /* 固定一行 tag 高度（含边框与行高余量），无标签也占位 */
  height: 1.65rem;
  overflow: hidden;
}

.projects-card__tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  flex-shrink: 0;
  max-width: 7.5rem;
  height: 1.5rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 6px;
  background: var(--efu-secondbg);
  color: var(--efu-secondtext);
  padding: 0 0.45rem;
  font-size: 0.72rem;
  line-height: 1;
  text-decoration: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.projects-card__tag:hover {
  color: var(--milk-ink);
  background: var(--efu-ahoverbg);
}

.projects-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-top: auto;
  padding-top: 0.35rem;
}

.projects-card__action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.85rem;
  padding: 0.2rem 0.7rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 8px;
  background: var(--efu-card-bg);
  color: var(--milk-ink);
  font-size: 0.8rem;
  text-decoration: none;
  transition:
    border-color 0.15s ease,
    background 0.15s ease,
    color 0.15s ease;
}

.projects-card__action:hover {
  border-color: color-mix(in srgb, var(--milk-ink) 28%, var(--efu-card-border));
  background: var(--efu-ahoverbg);
}

.projects-card__action.is-primary {
  border-color: transparent;
  background: var(--milk-ink);
  color: #fff;
}

.projects-card__action.is-primary:hover {
  opacity: 0.9;
  background: var(--milk-ink);
  color: #fff;
}

@media (max-width: 639px) {
  .projects-card__cover {
    aspect-ratio: 16 / 7;
  }

  .projects-card__body {
    padding: 0.95rem 1rem 1rem;
  }

  .projects-card__title {
    font-size: 1.08rem;
  }

  .projects-card__desc {
    font-size: 0.9rem;
  }
}
</style>
