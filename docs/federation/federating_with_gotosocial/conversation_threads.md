# Conversation Threads

Due to the nature of decentralization and federation, it is practically impossible for any one server on the fediverse to be aware of every post in a given conversation thread.

With that said, it is possible to do 'best effort' dereferencing of threads, whereby remote replies are fetched from one server onto another, to try to more fully flesh out a conversation.

GoToSocial does this by iterating up and down the thread of a conversation, pulling in remote statuses where possible.

Let's say we have two accounts: `local_account` on `our.server`, and `remote_1` on `remote.1`.

In this scenario, `local_account` follows `remote_1`, so posts from `remote_1` show up in the home timeline of `local_account`.

Now, `remote_1` boosts/reblogs a post from a third account, `remote_2`, residing on server `remote.2`.

`local_account` does not follow `remote_2`, and neither does anybody else on `our.server`, which means that `our.server` has not seen this post by `remote_2` before.

![A diagram of the conversation thread, showing the post from remote_2, and possible ancestor and descendant posts](../../assets/diagrams/conversation_thread.png)

What GoToSocial will do now, is 'dereference' the post by `remote_2` to check if it is part of a thread and, if so, whether any other parts of the thread can be obtained.

GtS begins by checking the `inReplyTo` property of the post, which is set when a post is a reply to another post. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-inreplyto). If `inReplyTo` is set, GoToSocial derefences the replied-to post. If *this* post also has an `inReplyTo` set, then GoToSocial dereferences that too, and so on.

Once all of these **ancestors** of a status have been retrieved, GtS will begin working down through the **descendants** of posts.

It does this by checking the `replies` property of a derefenced post, and working through replies, and replies of replies. [See here](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-replies).

This process of thread dereferencing will likely involve making multiple HTTP calls to different servers, especially if the thread is long and complicated.

The end result of this dereferencing is that, assuming the reblogged post by `remote_2` was part of a thread, then `local_account` should now be able to see posts in the thread when they open the status on their home timeline. In other words, they will see replies from accounts on other servers (who they may not have come across yet), in addition to any previous and next posts in the thread as posted by `remote_2`.

This gives `local_account` a more complete view on the conversation, as opposed to just seeing the reblogged post in isolation and out of context. It also gives `local_account` the opportunity to discover new accounts to follow, based on replies to `remote_2`.
