# Naming / Repository Name

Status: **draft**

## Proposed name: Text Eater

`Text Eater` fits the current product direction better than `Open LLM Learner` because the project is no longer LLM-first.

The product works with text-centered tasks such as:
- generation
- classification
- extraction
- ranking
- retrieval
- transformation
- code editing
- judgment
- embeddings
- LLM training

The name communicates that the system consumes text and turns it into useful behavior without implying that every model must be an LLM.

## Repository name

Preferred repository name:

`text-eater`

The repository name appears available in the current GitHub account at the time this note was written.

## Internal IDs

Repository/product naming and protocol/schema namespace are separate decisions.

Existing format identifiers currently use names such as:

`open-llm-learner.task`

Because the project is still in design stage, changing those identifiers is possible, but it should be done as one deliberate migration rather than piecemeal.

Recommended sequence:
1. decide product/repository name;
2. rename repository;
3. decide whether JSON format IDs / official MOD IDs also move to a `text-eater.*` namespace;
4. update docs and schemas in one migration commit.

Until that migration is performed, `Text Eater` is the preferred working product name in design discussions while existing machine identifiers remain unchanged.