# 端到端用户工作流测试
# Phase 14: Task 30.4 - End-to-End User Workflow Testing

Feature: End-to-End User Workflow Testing
  As a user
  I want to perform complete workflows seamlessly
  So that I can accomplish my bookmark management goals

  Background:
    Given the bookmark synchronization service is running
    And all browser extensions are available
    And the web interface is accessible

  @e2e-testing @user-onboarding
  Scenario: End-to-End User Registration and Login
    Given I am a new user visiting the service
    When I register a new account with valid credentials
    And I verify my email address
    And I log in with my credentials
    Then I should be successfully authenticated
    And I should see the main dashboard
    And my user profile should be created
    And welcome notifications should be displayed

  @e2e-testing @bookmark-management
  Scenario: Complete Bookmark Management Workflow
    Given I am logged in as a registered user
    When I create a new bookmark from a webpage
    And I add the bookmark to a collection
    And I edit the bookmark details
    And I search for the bookmark
    And I organize bookmarks into folders
    Then all bookmark operations should work seamlessly
    And changes should be reflected immediately
    And search should find the correct bookmarks
    And folder organization should be preserved

  @e2e-testing @cross-browser-sync
  Scenario: Cross-Browser Synchronization Workflow
    Given I have the browser extension installed on Chrome
    And I have the browser extension installed on Firefox
    When I save a bookmark in Chrome extension
    And I wait for synchronization to complete
    And I open Firefox extension
    Then the bookmark should appear in Firefox
    And metadata should be identical across browsers
    And sync status should show as successful
    And conflict resolution should work automatically

  @e2e-testing @offline-online
  Scenario: Offline to Online Synchronization Workflow
    Given I am using the browser extension
    And I go offline by disconnecting from internet
    When I create several bookmarks while offline
    And I edit existing bookmarks while offline
    And I reconnect to internet
    Then offline changes should sync automatically
    And conflict resolution should handle duplicates
    And all data should be consistent
    And sync status should update correctly

  @e2e-testing @sharing-collaboration
  Scenario: Bookmark Sharing and Collaboration Workflow
    Given I am logged in with bookmarks and collections
    When I share a collection publicly
    And another user discovers the shared collection
    And they fork the collection to their account
    And they add new bookmarks to their fork
    Then sharing should work correctly
    And collection discovery should function properly
    And forking should create independent copies
    And collaboration features should be accessible

  @e2e-testing @import-export
  Scenario: Data Import and Export Workflow
    Given I am logged in to the service
    When I import bookmarks from Chrome browser
    And I import bookmarks from Firefox browser
    And I organize the imported bookmarks
    And I export all bookmarks to JSON format
    Then all imports should preserve folder structure
    And duplicate detection should work correctly
    And exported data should be complete and valid
    And reimporting exported data should work

  @e2e-testing @search-discovery
  Scenario: Advanced Search and Discovery Workflow
    Given I have a large collection of bookmarks
    When I search using text queries
    And I search using Chinese language terms
    And I use advanced filters and faceted search
    And I browse trending and recommended bookmarks
    Then search results should be accurate and fast
    And Chinese language search should work properly
    And filtering should narrow results effectively
    And recommendations should be relevant

  @e2e-testing @mobile-responsive
  Scenario: Mobile Responsive Interface Workflow
    Given I access the service on a mobile device
    When I navigate through the mobile interface
    And I perform bookmark operations on mobile
    And I switch between mobile and desktop
    Then mobile interface should be fully functional
    And responsive design should adapt correctly
    And touch interactions should work properly
    And data should sync across devices

  @e2e-testing @performance-user-experience
  Scenario: Performance and User Experience Workflow
    Given I am using the service with a large dataset
    When I perform various operations
    And I measure page load times
    And I test with slow network conditions
    Then response times should be acceptable
    And user interface should remain responsive
    And loading indicators should provide feedback
    And error handling should be user-friendly