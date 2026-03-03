# Kubeflow’s Release Process Addendum - Release Blog Post and Presentation


# Purpose

This document reviews the Kubeflow Community’s process for a blog post and a recorded presentation, which are delivered with each Kubeflow release.   The blog post and presentation are used to increase the Community’s understanding of new features in a Kubeflow release.  


## Initial note

Both of these deliveries (blog and presentation) have much of the same content.   The presentation can usually be completed in one shot with some quick edits.  It provides valuable information that can be used in the blog.  The blog requires more precision, and therefore takes multiple iterations.  It is suggested to start the blog before the presentation.


# Blog Post


## Blog Overview



* Publish a blog of 1,000 words or less that reviews the highlights of new release
* The blog will promote themes and the release’s top features.  The blog is a summary of the release, and more detailed information is provided in links to GitHub Issues, Pull Requests (PR), additional (Working Group or User) blogs, repos or other resources.
* The post uses the GitHub PR fast pages process, which documented here, ​​https://github.com/kubeflow/blog#blog-submission-instructions
    * This PR is an example of the 1.4 blog post, [https://github.com/kubeflow/blog/pull/109](https://github.com/kubeflow/blog/pull/109) 


## Community Resources needed



* Product Manager - 20 hours
    * Manages the blog’s creation and release process,  creates initial version, asks and incorporates edits, keeps process moving, provides updates to Community
* GitHub specialist - 10 hours
    * Creates PR, helps to check content, incorporate edits to PR
* Content creators, editors and approvers - 2 hours
    * Working Group leads provide content, review text, approve via /lgtm in PR
* Final Approver - 4 hours
    * Kubeflow Steering Committee lead


## Example 

An example of a release blog post is here: https://blog.kubeflow.org/kubeflow-1.4-release/

Note - A sample of the work activities and timeline are detailed in a subsequent section of this document.


# Presentation


## Presentation overview

30 minute slide presentation recorded and presented by the Kubeflow Community Product Management, the Release Team and the Working Group Leads.   The presentation should be posted on the [Kubeflow Youtube channel](https://www.youtube.com/channel/UCNSJwHkt2f_cP_UN2q-yJlw).


## Community Resources needed



* Product Manager - 20 hours
    * Manages the process, creates initial presentation outline, asks working groups for content and speakers, reviews and incorporates edits, keeps, provides updates to Community, keeps the process moving
* Content creators, editors and approvers - 3 hours
    * Working Group leads provide content, review text, approve via /lgtm in PR
* Video Editor
    * We need a pro at editing videos (that are currently created via Zoom)
* Final Approver - 2 hours
    * Kubeflow Steering Committee lead


## Example: 

An example of the recorded presentation is here: [https://www.youtube.com/watch?v=gG61gHw4J14](https://www.youtube.com/watch?v=gG61gHw4J14)

An example of the slide is available here: [https://docs.google.com/presentation/d/1S7FRa8DI5uymUuPv56ioC37QXgb7kxSC5p7pp0gMd9U/edit#slide=id.gba45ef4eaa_0_1651](https://docs.google.com/presentation/d/1S7FRa8DI5uymUuPv56ioC37QXgb7kxSC5p7pp0gMd9U/edit#slide=id.gba45ef4eaa_0_1651)


# Sample Activity Timeline

This timeline shows potential activities for the blog and presentation, starting from 35 days from the postings.


<table>
  <tr>
   <td>Blog
   </td>
   <td>Presentation
   </td>
  </tr>
  <tr>
   <td>35 days out
<p>
Collect features (including PR# links) from Kubeflow GitHub repositories and roadmaps
<p>
Kubeflow and Kubeflow component roadmaps: <a href="https://github.com/kubeflow/kubeflow/blob/master/ROADMAP.md">https://github.com/kubeflow/kubeflow/blob/master/ROADMAP.md</a> 
<p>
Ask working group leads for a list of top features during working group calls
<p>
Review, group and prioritize features
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>28 days out
<p>
Create themes and top feature list
<p>
Ask working groups for their input
<p>
Create blog outline in google doc (please find a sample release outline below)
<p>
Sample Blog Outline
<p>
Product Management update on themes, top features, process improvements
<p>
Individual Working Groups updates (Manifests, Notebooks, Katib, Training, KFPipelines, KServe)
<p>
Note - you might ask for top 3 feature names, short descriptions, benefits, PRs.   
<p>
Distributions
<p>
Tutorials
<p>
What’s next
<p>
Getting Involved
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>21 days out
<p>
Write 1st version in google doc
<p>
Review 1st version with a few folks in Community i.e. Release Manager, a WG lead
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>
   </td>
   <td>17 days out
<p>
Create issue in Kubeflow/Community repo for Release video presentation
<p>
Create themes and top feature list, ask working groups for their input
<p>
Create presentation in google slides
<p>
Sample Release Outline (similar to blog)
<p>
Product Management update on themes, top features, process improvements
<p>
Working Groups updates - 1 or 2 slides per WG
<p>

    Provide a format for slides - i.e. Info on WG Meeting (Time, Notes, Link) along with Top 3 feature names, short description, benefit and PR 
<p>
Distributions
<p>
Tutorials
<p>
What’s next
<p>
Getting Involved
   </td>
  </tr>
  <tr>
   <td>14 days
<p>
Write 2nd version of blog
<p>
Share 2nd version with the Working Group Leads and Release team in google doc.
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>
   </td>
   <td>12 days out
<p>
Schedule date for presentation,
<p>
Identify speakers for recorded presentation
   </td>
  </tr>
  <tr>
   <td>
   </td>
   <td>11 days out
<p>
Review Working Group edits
<p>
Create and distribute 2rd version to Working Groups
<p>
Ask for final comments
   </td>
  </tr>
  <tr>
   <td>9 days out
<p>
Create 3rd version in google doc
<p>
Incorporate comments
<p>
Confirm the Links in doc
<p>
Note - Inputting links in a PR does take time and the number of links should be kept to a reasonable level.  Please confirm that the links are correct before creating the PR.
<p>
Move content from google doc to PR
<p>
Create PR using fast pages,https://github.com/kubeflow/blog#blog-submission-instructions
<p>
Ask for comments, edits from Working Groups leads on PR
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>7 days out
<p>
Update PR based on comments
<p>
Ask for final comments / final reviews
<p>
Confirm the Links
<p>
Ask Working Group leads for /lgtm
<p>
Alert Steering Committee that near final version is close
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>
   </td>
   <td>6 days out
<p>
Record presentation
   </td>
  </tr>
  <tr>
   <td>
   </td>
   <td>5 days out
<p>
Review presentation, ask for any comments/edits, re-recordings
<p>
Provide to Video Pro and complete splice in edits.
<p>
<strong>Tools</strong>
<p>
Quicktime
<p>
Premiere pro
<p>
Handbrake
<p>
Sample Time - 1-3 days
   </td>
  </tr>
  <tr>
   <td>4 days out
<p>
Add in anything learned from presentation
   </td>
   <td>4 days out
<p>
Complete final edits and review
<p>
Provide /lgtm from content providers and approvers 
   </td>
  </tr>
  <tr>
   <td>3 days out
<p>
Get final /ltgm from stakeholders
<p>
Steering Committee adds Final edits
   </td>
   <td>3 days out 
<p>
Submit to Steering committee for posting on Kubeflow channel on YouTube
   </td>
  </tr>
  <tr>
   <td>Day of Post
<p>
Steering Committee approves and posts
<p>
Ask all Resources to review for any corrections
<p>
Have team ready for an quick updates
<p>
Promote on Kubeflow slack and on Kubeflow-discuss
   </td>
   <td>Day of Post
<p>
Highlight in presentation on Kubeflow-discuss mailing list (<a href="mailto:kubeflow-discuss@googlegroups.com">kubeflow-discuss@googlegroups.com</a> and on Kubeflow-slack (kubeflow.slack.com) in general channel
   </td>
  </tr>
</table>
