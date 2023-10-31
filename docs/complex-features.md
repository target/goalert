# The Phased Approach to Feature Development

In the course of our development, we recognize that not all features are born equal. Some are more complex, multifaceted, or novel. For these "not a get-it-right-the-first-time-features", we advocate for a phased approach.

## Why the Phased Approach?

1. **Safe Exploration:** It provides a safety net for developers to explore, make mistakes, and learn.
2. **Enhanced Collaboration:** It encourages early sharing and iterative feedback.
3. **Quality End Product:** It ensures that our end product benefits from iterative insights.

## The Steps

1. **Proposal & Discussion:** Every feature starts with a proposal. It's essential to have buy-in and a clear understanding of what we're building.

2. **Prototype to Learn:** The initial step is to build a working prototype. This isn't about perfection; it's about exploration and understanding.

3. **Iterative Development:** Based on feedback and learnings from the prototype, we refine and develop the feature further.

4. **Refinement (Optional):** Depending on the complexity, we might go through another round of refinement. 

5. **Review & Merge:** As with all our features, the final step is a thorough review followed by a merge.

## A Note for Developers

Whether you're a seasoned developer with our organization or a new open-source contributor, remember: It's okay not to get it right the first time. Our goal is continuous improvement and learning. This phased approach is a tool to help us achieve that while delivering outstanding features.

## Handling Larger Features

For features or changes that would span over 500 lines or are inherently complex, we adopt an incremental delivery approach. This not only manages risk but also promotes modular design and continuous integration. This isnt' a hard rule, but a guideline to help us deliver features effectively.

### Strategies for Incremental Delivery:

1. **Experimental Flags:** Use experimental flags to lock unfinished or experimental parts. This way, we can merge code into the mainline without affecting the overall functionality until the feature is complete. Remember, tests are required to remove these flags, ensuring quality control.

2. **Separate Backend and Frontend:** When necessary, split backend/API changes from frontend/UI changes to modularize the feature and allow for parallel development.

3. **Granular Task Breakdown:** Even within specific domains like backend or frontend, you may choose to further break down tasks for manageable PRs. For example, in backend development, instead of a single PR for CRUD operations, consider separate PRs for Create & Delete and another for List & Update. Similarly, in frontend, implement listing items first, followed by Create & Edit functionalities in subsequent PRs (since those are generally handled from the list).

4. **Modular Design:** Aim to design features in a way that they can be built, tested, and delivered in smaller chunks. This reduces the gravity of change and aids in quicker reviews and iterations.

5. **Testable Pull Requests:** Each PR should introduce a verifiable change. Including tests for that specific change ensures that we're always moving forward with confidence.

### Benefits:

- **Reduced Risk:** Smaller, incremental changes are easier to test, review, and debug.
  
- **Avoid Churn:** Incremental delivery reduces merge conflicts and integration issues.
  
- **Improved Collaboration:** Smaller changes allow for quicker feedback cycles and more collaborative development.

Remember, the goal isn't just to deliver features; it's to deliver them effectively, efficiently, and with the highest quality.
